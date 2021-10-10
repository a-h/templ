package testscriptusage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `` +
	`<script type="text/javascript">function withParameters(a, b, c){console.log(a, b, c);}function withoutParameters(){alert("hello");}</script>` +
	`<button onClick="withParameters(&#34;test&#34;,&#34;A&#34;,123)" onMouseover="withoutParameters()" type="button">A</button>` +
	`<button onClick="withParameters(&#34;test&#34;,&#34;B&#34;,123)" onMouseover="withoutParameters()" type="button">B</button>` +
	`<button onMouseover="console.log(&#39;mouseover&#39;)" type="button">Button C</button>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := ThreeButtons().Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
	fmt.Println(w.String())
}
