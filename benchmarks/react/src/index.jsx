import * as React from 'react'
import * as Server from 'react-dom/server'
import { performance } from 'node:perf_hooks';

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

// Warm up.
for (let i = 0; i < 1000; i++) {
        Server.renderToString(component(p));
}

// Benchmark.
const iterations = 100000;
const start = performance.now();
for (let i = 0; i < iterations; i++) {
        Server.renderToString(component(p));
}
const elapsed = performance.now() - start;
const opsPerSec = Math.round(iterations / (elapsed / 1000));
const nsPerOp = Math.round((elapsed / iterations) * 1e6);

console.log(`Render test x ${opsPerSec.toLocaleString()} ops/sec`);
console.log(`${nsPerOp.toLocaleString()} ns/op`);
