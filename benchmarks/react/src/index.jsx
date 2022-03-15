var ReactDOMServer = require('react-dom/server');
var Benchmark = require('benchmark');

function element(p) {
        return <div>
                <h1>{p.Name}</h1>
                <div style={{ fontFamily: "sans-serif" }} id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;">
                        <div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div>
                </div>
                <hr noshade /><hr optionA optionB optionC="other" /><hr noshade />
        </div>;
}

function streamToString(stream) {
        const chunks = [];
        return new Promise((resolve, reject) => {
                stream.on('data', (chunk) => chunks.push(Buffer.from(chunk)));
                stream.on('error', (err) => reject(err));
                stream.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')));
        })
}

let p = {
        Name: "Luiz Bonfa",
        Email: "luiz@example.com",
};

async function test() {
        const stream = ReactDOMServer.renderToNodeStream(element(p));
        return await streamToString(stream);
}

//test()
//.then(html => console.log(html))
//.catch(err => console.log(err));

// Benchmark.
// 156,735 operations per second.
// There are 1,000,000,000 nanoseconds in a second.
// So this is 6,380 ns per operation.
// The templ equivalent is 1,167 ns per operation.
var suite = new Benchmark.Suite;
suite.add('Render test', function() {
        test().then(() => null)
})
        .on('cycle', function(event) {
                console.log(String(event.target));
        })
        .on('complete', function() {
                console.log('Fastest is ' + this.filter('fastest').map('name'));
        })
        .run({ 'async': true });
