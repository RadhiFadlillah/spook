# Spook

Spook is a simple CMS for creating a simple blog. Built using Go and Vue.js. Inspired by Ghost and Hugo, but smaller and has less feature.

## Features

- Minimalistic and only has the bare necessary function for creating and managing a blog.
- Easy to install, and can be installed on any server;
- Doesn't require any external dependency;
- Has search feature;
- Easy and flexible to theme;

## Technology Stack

Spook uses Go for back-end, Vue.js for front-end and SQLite3 for database :

- Go is used because it can be cross compiled into multiple platform. The compile result is a single binary executable, which make it easy to install and distibute.
- Vue.js is used because it's dead simple to use. It still can be installed by including `<script>` tag, unlike its rival like Angular or React. For modern web development, JS bundler might be the way to go. However, for some reason I still can't wrap my head around it.
- SQLite3 is used because it's light, fast, portable, and available in almost all platform. While it doesn't support concurrent write, it supports concurrent read which make it perfect for small blog.

## License

Spook is distributed under Apache-2.0 License. Basically, it means you can do what you like with the software. However, if you modify it, you have to include the license and notices, and state what did you change.