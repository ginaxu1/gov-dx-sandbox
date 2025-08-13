const http = require('http');

// Read environment variables
const PORT = process.env.PORT || 3000;
const GREETING = process.env.GREETING || 'Hello, world!';

const server = http.createServer((req, res) => {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ message: GREETING }));
});

server.listen(PORT, () => {
    console.log(`Mock server running on port ${PORT}`);
});