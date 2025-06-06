package storybook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/sumdb/dirhash"

	_ "embed"

	"github.com/a-h/templ"
	"github.com/rs/cors"
)

type Storybook struct {
	// Path to the storybook-server directory, defaults to ./storybook-server.
	Path string
	// RoutePrefix is the prefix of HTTP routes, e.g. /prod/
	RoutePrefix string
	// Config of the Stories.
	Config map[string]*Conf
	// Handlers for each of the components.
	Handlers map[string]http.Handler
	// Handler used to serve Storybook, defaults to filesystem at ./storybook-server/storybook-static.
	StaticHandler      http.Handler
	Header             string
	Server             http.Server
	Log                *slog.Logger
	AdditionalPrefixJS string
}

type StorybookConfig func(*Storybook)

func WithServerAddr(addr string) StorybookConfig {
	return func(sb *Storybook) {
		sb.Server.Addr = addr
	}
}

func WithHeader(header string) StorybookConfig {
	return func(s *Storybook) {
		s.Header = header
	}
}

func WithPath(path string) StorybookConfig {
	return func(sb *Storybook) {
		sb.Path = path
	}
}

// WithAdditionalPreviewJS / WithAdditionalPreviewJS allows to add content to the generated .storybook/preview.js file.
// For example this can be used to include custom CSS.
func WithAdditionalPreviewJS(content string) StorybookConfig {
	return func(sb *Storybook) {
		sb.AdditionalPrefixJS = content
	}
}

func New(conf ...StorybookConfig) *Storybook {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	sh := &Storybook{
		Path:     "./storybook-server",
		Config:   map[string]*Conf{},
		Handlers: map[string]http.Handler{},
		Log:      logger,
	}
	sh.Server = http.Server{
		Handler: sh,
		Addr:    ":60606",
	}
	for _, sc := range conf {
		sc(sh)
	}

	// Depends on the correct Path, so must be set after additional config
	sh.StaticHandler = http.FileServer(http.Dir(path.Join(sh.Path, "storybook-static")))

	return sh
}

func (sh *Storybook) AddComponent(name string, componentConstructor any, args ...Arg) *Conf {
	// TODO: Check that the component constructor is a function that returns a templ.Component.
	c := NewConf(name, args...)
	sh.Config[name] = c
	h := NewHandler(name, componentConstructor, args...)
	sh.Handlers[name] = h
	return c
}

func (sh *Storybook) Build(ctx context.Context) (err error) {
	// Download Storybook to the directory required.
	sh.Log.Info("Installing storybook.")
	err = sh.installStorybook()
	if err != nil {
		return
	}
	if err = ctx.Err(); err != nil {
		return
	}

	// Copy the config to Storybook.
	sh.Log.Info("Configuring storybook.")
	configHasChanged, err := sh.configureStorybook()
	if err != nil {
		return
	}
	if err = ctx.Err(); err != nil {
		return
	}

	// Execute a static build of storybook if the config has changed.
	if configHasChanged {
		sh.Log.Info("Config not present, or has changed, rebuilding storybook.")
		err = sh.buildStorybook()
		if err != nil {
			return
		}
	} else {
		sh.Log.Info("Storybook is up-to-date, skipping build step.")
	}
	if err = ctx.Err(); err != nil {
		return
	}

	return
}

func (sh *Storybook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sbh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, path.Join(sh.RoutePrefix, "/storybook_preview/")) {
			sh.previewHandler(w, r)
			return
		}
		sh.StaticHandler.ServeHTTP(w, r)
	})
	cors.Default().Handler(sbh).ServeHTTP(w, r)
}

func (sh *Storybook) ListenAndServeWithContext(ctx context.Context) (err error) {
	err = sh.Build(ctx)
	if err != nil {
		return
	}
	go func() {
		sh.Log.Info("Starting Go server", slog.String("address", sh.Server.Addr))
		err = sh.Server.ListenAndServe()
	}()
	<-ctx.Done()
	// Close the Go server.
	_ = sh.Server.Close()
	return err
}

func (sh *Storybook) previewHandler(w http.ResponseWriter, r *http.Request) {
	prefix := path.Join(sh.RoutePrefix, "/storybook_preview/")
	if !strings.HasPrefix(r.URL.Path, prefix) {
		sh.Log.Warn("URL does not match preview prefix", slog.String("url", r.URL.String()))
		http.NotFound(w, r)
		return
	}

	name, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, prefix))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to unescape URL: %v", err), http.StatusBadRequest)
		return
	}
	if name == "" {
		sh.Log.Warn("URL does not contain component name", slog.String("url", r.URL.String()))
		http.NotFound(w, r)
		return
	}
	name = strings.TrimPrefix(name, "/")

	h, found := sh.Handlers[name]
	if !found {
		sh.Log.Info("Component name not found", slog.String("name", name), slog.String("url", r.URL.String()), slog.Any("available", keysOfMap(sh.Handlers)))
		http.NotFound(w, r)
		return
	}
	h.ServeHTTP(w, r)
}

func keysOfMap[K comparable, V any](handler map[K]V) (keys []K) {
	keys = make([]K, len(handler))
	var i int
	for k := range handler {
		keys[i] = k
		i++
	}
	return keys
}

//go:embed _package.json
var packageJSON string

func (sh *Storybook) installStorybook() (err error) {
	_, err = os.Stat(sh.Path)
	if err == nil {
		sh.Log.Info("Storybook already installed, Skipping installation.")
		return
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(sh.Path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("templ-storybook: error creating @storybook/server directory: %w", err)
		}
		err = os.WriteFile(filepath.Join(sh.Path, "package.json"), []byte(packageJSON), 0644)
		if err != nil {
			return fmt.Errorf("templ-storybook: error writing package.json: %w", err)
		}
	}
	var cmd exec.Cmd
	cmd.Dir = sh.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Path, err = exec.LookPath("npx")
	if err != nil {
		return fmt.Errorf("templ-storybook: cannot install storybook, cannot find npx on the path, check that Node.js is installed: %w", err)
	}
	cmd.Args = []string{"npx", "sb", "init", "--features", "docs", "test", "-t", "server", "--no-dev"}
	return cmd.Run()
}

func (sh *Storybook) configureStorybook() (configHasChanged bool, err error) {
	// Delete template/existing files in the stories directory.
	storiesDir := filepath.Join(sh.Path, "stories")
	before, err := dirhash.HashDir(storiesDir, "/", dirhash.DefaultHash)
	if err != nil && !os.IsNotExist(err) {
		return configHasChanged, err
	}
	if err = os.RemoveAll(storiesDir); err != nil {
		return configHasChanged, err
	}
	if err := os.Mkdir(storiesDir, os.ModePerm); err != nil {
		return configHasChanged, err
	}
	// Create new *.stories.json files.
	for _, c := range sh.Config {
		name := filepath.Join(sh.Path, fmt.Sprintf("stories/%s.stories.json", c.Title))
		f, err := os.Create(name)
		if err != nil {
			return configHasChanged, fmt.Errorf("failed to create config file to %q: %w", name, err)
		}
		err = json.NewEncoder(f).Encode(c)
		if err != nil {
			return configHasChanged, fmt.Errorf("failed to write JSON config to %q: %w", name, err)
		}
	}
	after, err := dirhash.HashDir(storiesDir, "/", dirhash.DefaultHash)
	if err != nil {
		return configHasChanged, fmt.Errorf("failed to hash directory %q: %w", storiesDir, err)
	}
	configHasChanged = before != after
	// Configure storybook Preview URL.
	err = os.WriteFile(filepath.Join(sh.Path, ".storybook/preview.js"), []byte(fmt.Sprintf("%s\n%s", sh.AdditionalPrefixJS, previewJS)), os.ModePerm)
	if err != nil {
		return
	}
	// Configure preview-head.html
	err = os.WriteFile(filepath.Join(sh.Path, ".storybook/preview-head.html"), []byte(sh.Header), os.ModePerm)
	return
}

var previewJS = `
// Customise fetch so that it uses a relative URL.
const fetchStoryHtml = async (url, path, params, context) => {
  const qs = new URLSearchParams(params);
  const response = await fetch("/storybook_preview/" + path + "?" + qs.toString());
  return response.text();
};

export const parameters = {
  server: {
    url: "http://localhost/storybook_preview", // Ignored by fetchStoryHtml.
    fetchStoryHtml,
  },
};
`

func (sh *Storybook) buildStorybook() (err error) {
	var cmd exec.Cmd
	cmd.Dir = sh.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Path, err = exec.LookPath("npm")
	if err != nil {
		return fmt.Errorf("templ-storybook: cannot run storybook, cannot find npm on the path, check that Node.js is installed: %w", err)
	}
	cmd.Args = []string{"npm", "run", "build-storybook"}
	return cmd.Run()
}

func NewHandler(name string, f any, args ...Arg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		argv := make([]any, len(args))
		q := r.URL.Query()
		for i, arg := range args {
			argv[i] = arg.Get(q)
		}
		component, err := executeTemplate(name, f, argv)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(component).ServeHTTP(w, r)
	})
}

func executeTemplate(name string, fn any, values []any) (output templ.Component, err error) {
	v := reflect.ValueOf(fn)
	t := v.Type()
	argv := make([]reflect.Value, t.NumIn())
	if len(argv) != len(values) {
		err = fmt.Errorf("templ-storybook: component %s expects %d argument, but %d were provided", fn, len(argv), len(values))
		return
	}
	for i := range argv {
		argv[i] = reflect.ValueOf(values[i])
	}
	result := v.Call(argv)
	if len(result) != 1 {
		err = fmt.Errorf("templ-storybook: function %s must return a templ.Component", name)
		return
	}
	output, ok := result[0].Interface().(templ.Component)
	if !ok {
		err = fmt.Errorf("templ-storybook: result of function %s is not a templ.Component", name)
		return
	}
	return output, nil
}

func NewConf(title string, args ...Arg) *Conf {
	c := &Conf{
		Title: title,
		Parameters: StoryParameters{
			Server: map[string]any{
				"id": title,
			},
		},
		Args:     NewSortedMap(),
		ArgTypes: NewSortedMap(),
		Stories:  []Story{},
	}
	for _, arg := range args {
		c.Args.Add(arg.Name, arg.Value)
		c.ArgTypes.Add(arg.Name, map[string]any{
			"control": arg.Control,
		})
	}
	c.AddStory("Default")
	return c
}

func (c *Conf) AddStory(name string, args ...Arg) {
	m := NewSortedMap()
	for _, arg := range args {
		m.Add(arg.Name, arg.Value)
	}
	c.Stories = append(c.Stories, Story{
		Name: name,
		Args: m,
	})
}

// Controls for the configuration.
// See https://storybook.js.org/docs/react/essentials/controls
type Arg struct {
	Name    string
	Value   any
	Control any
	Get     func(q url.Values) any
}

func ObjectArg(name string, value any, valuePtr any) Arg {
	return Arg{
		Name:    name,
		Value:   value,
		Control: "object",
		Get: func(q url.Values) any {
			err := json.Unmarshal([]byte(q.Get(name)), valuePtr)
			if err != nil {
				return err
			}
			return reflect.Indirect(reflect.ValueOf(valuePtr)).Interface()
		},
	}
}

func TextArg(name, value string) Arg {
	return Arg{
		Name:    name,
		Value:   value,
		Control: "text",
		Get: func(q url.Values) any {
			return q.Get(name)
		},
	}
}

func BooleanArg(name string, value bool) Arg {
	return Arg{
		Name:    name,
		Value:   value,
		Control: "boolean",
		Get: func(q url.Values) any {
			return q.Get(name) == "true"
		},
	}
}

type IntArgConf struct{ Min, Max, Step *int }

func IntArg(name string, value int, conf IntArgConf) Arg {
	control := map[string]any{
		"type": "number",
	}
	if conf.Min != nil {
		control["min"] = conf.Min
	}
	if conf.Max != nil {
		control["max"] = conf.Max
	}
	if conf.Step != nil {
		control["step"] = conf.Step
	}
	arg := Arg{
		Name:    name,
		Value:   value,
		Control: control,
		Get: func(q url.Values) any {
			i64, err := strconv.ParseInt(q.Get(name), 10, 64)
			if err != nil || i64 < math.MinInt || i64 > math.MaxInt {
				return 0
			}
			return int(i64)
		},
	}
	return arg
}

func FloatArg(name string, value float64, min, max, step float64) Arg {
	return Arg{
		Name:  name,
		Value: value,
		Control: map[string]any{
			"type": "number",
			"min":  min,
			"max":  max,
			"step": step,
		},
		Get: func(q url.Values) any {
			i, _ := strconv.ParseFloat(q.Get(name), 64)
			return i
		},
	}
}

type Conf struct {
	Title      string          `json:"title"`
	Parameters StoryParameters `json:"parameters"`
	Args       *SortedMap      `json:"args"`
	ArgTypes   *SortedMap      `json:"argTypes"`
	Stories    []Story         `json:"stories"`
}

type StoryParameters struct {
	Server map[string]any `json:"server"`
}

func NewSortedMap() *SortedMap {
	return &SortedMap{
		m:        new(sync.Mutex),
		internal: map[string]any{},
		keys:     []string{},
	}
}

type SortedMap struct {
	m        *sync.Mutex
	internal map[string]any
	keys     []string
}

func (sm *SortedMap) Add(key string, value any) {
	sm.m.Lock()
	defer sm.m.Unlock()
	sm.keys = append(sm.keys, key)
	sm.internal[key] = value
}

func (sm *SortedMap) MarshalJSON() (output []byte, err error) {
	sm.m.Lock()
	defer sm.m.Unlock()
	b := new(bytes.Buffer)
	b.WriteRune('{')
	enc := json.NewEncoder(b)
	for i, k := range sm.keys {
		err = enc.Encode(k)
		if err != nil {
			return
		}
		_, err = b.WriteRune(':')
		if err != nil {
			return
		}
		err = enc.Encode(sm.internal[k])
		if err != nil {
			return
		}
		if i < len(sm.keys)-1 {
			_, err = b.WriteRune(',')
			if err != nil {
				return
			}
		}
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

type Story struct {
	Name string     `json:"name"`
	Args *SortedMap `json:"args"`
}
