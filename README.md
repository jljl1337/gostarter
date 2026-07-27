# gostarter

- Batteries included, offers cron jobs, backups and more.
- Configurable, either full on setup, or just basic tables, or even build from scratch.
- Infrastructure independent, works with both monolith and microservice architecture.
- Compatible with SQLite and PostgreSQL.

## What is gostarter?

gostarter is a package aiming to provide useful components to create a monolith backend server. That said, gostarter can also be configured to act as just a Restful API server or a sidecar application if you are adopting a microservice architecture.

## Why not PocketBase (or other alternatives)?

### PocketBase

PocketBase assumes you to use the record and collection operations instead of messing with raw SQL queries, which works very well for many cases. That said, you may still want to go around the API and extend the app for your use case from time to time. If this happens again and again, starting from scratch could be a better option.

If you also find the API of PocketBase a bit limiting, you can also explore the option of adopting this package, as it provides many commonly used components, while still offer enough flexibility for different kinds of applications.

All these aside, make sure to check out PocketBase before looking at gostarter, as it is still a more actively maintained and more mature product.

### Supabase

As for BaaS, Supabase is one of the very popular options. It is a much more complete solution compared to this Golang package. If you are building a prototype with possibility to serve millions of concurrent users, Supabase is a much more mature and suitable platform to develop your product.

That being said, you may realize the downside of Supabase when you try to self-host it on your machine. You have to clone the entire repository to install and update the whole infrastructure, making it hard to version control your backend. Also, the whole stack consume a lot of resources if you are just building a simple application. These and other caveats make self-hosting very far from an ideal solution for small to medium scale products, or internal applications.

## When not to use gostarter?

If you are building a backend with millions of concurrent users, avoid using this package in production workflow without any thorough stress test.

When you are developing prototype or facing requirements that change very frequently, as other options also provide interface for frontend to connect with the backend. This package only provides handy utilities to serve a Restful API, the frontend has to connect to the server by itself.

gostarter is very, very far from perfect, and is just a personal project with very limited resources, so please be aware of the shortcomings before using it.