const express = require('express');
const app = express();
app.use(express.static('.'));
app.get('/', (_, res) => res.redirect('/index.html'));
console.log('Listening on port 80');
app.listen(80);