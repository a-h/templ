import * as React from 'react'
import * as Server from 'react-dom/server'
import Benchmark from 'benchmark';

const component = (p) =>
        <div>
                <h1>{p.Name}</h1>
                <div style={{ fontFamily: "sans-serif" }} id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;">
                        <div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div>
                </div>
                <hr noshade /><hr optionA optionB optionC="other" /><hr noshade />
        </div>;

const p = {
        Name: "Luiz Bonfa",
        Email: "luiz@example.com",
};

// Benchmark.
// Outputs...
// Render test x 114,131 ops/sec ±0.27% (97 runs sampled)
// There are 1,000,000,000 nanoseconds in a second.
// 1,000,000,000ns / 114,131 ops = 8,757.5 ns per operation.
// The templ equivalent is 340 ns per operation.
const suite = new Benchmark.Suite;

const test = suite.add('Render test',
        () => Server.renderToString(component(p)))

test.on('cycle', (event) => {
        console.log(String(event.target));
});

test.run();
