# gostarter

[![Go Reference](https://pkg.go.dev/badge/github.com/jljl1337/gostarter.svg)](https://pkg.go.dev/github.com/jljl1337/gostarter)

- Batteries included, offers cron jobs, backups and more.
- Configurable, either full on setup, or table only, or even build from scratch.
- Infrastructure independent, from one binary for each sidecar to one binary for
  the entire backend.
- Compatible with SQLite and PostgreSQL.

## What is gostarter?

gostarter is a package aiming to provide useful components to create an easy to
manage monolith backend server, while still being possible to configured to act
as just a API server or a sidecar application if you want to separate your
backend into numerous processes or servers.

## Why not PocketBase (or other alternatives)?

### PocketBase

Before looking at gostarter, make sure to check out PocketBase, as it is still a
more actively maintained and more mature product. In many cases, PocketBase is a
much better choice overall.

By default, PocketBase assumes you to use the record and collection operations
instead of messing with raw SQL queries, which works very well for many cases.
Even though you can still execute raw SQL queries, you are then missing a lot of
features of using PocketBase as a framework. Also, if you want to separate the
backend into several binary, or use other SQL engine, you are out of luck.

### Supabase

As for BaaS, Supabase is one of the very popular options. It is a much more
complete solution compared to this Golang package. If you are building a
prototype with possibility to serve millions of concurrent users, Supabase is a
much more mature and suitable platform to develop your product.

That being said, you may realize the downside of Supabase when you try to
self-host it on your machine. You have to clone the entire repository to install
and update the whole infrastructure, making it hard to version control your
backend. Also, the whole stack consume a lot of resources if you are just
building a simple application. These and other caveats make self-hosting very
far from an ideal solution for small to medium scale products, or internal
applications.

## When not to use gostarter?

If you are building a backend with millions of concurrent users, avoid using
this package in production workflow without any thorough stress test.

When you are developing prototype or facing requirements that change very
frequently, as other options also provide interface for frontend to connect with
the backend. This package only provides handy utilities to serve a Restful API,
the frontend has to connect to the server by itself.

gostarter is very, very far from perfect, and is just a personal project with
very limited resources, so please be aware of the shortcomings before using it.
