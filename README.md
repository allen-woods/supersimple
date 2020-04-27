# supersimple

A research and iteration repo for learning `gqlgen`. This project implements aggregation in `MongoDB`, the `go-redis` package, and `APQ` for GraphQL.

This was a non-project research repo for cementing fundamentals of `gqlgen`, as well as exploring implementations of supporting technologies, such as `Redis`, `Apollo Client`, `APQ`, and `MongoDB`.

The repo has since evolved into a proof of concept that will serve as the basis for a project app that is currently untitled and in the concept phase.

# TODO:

- Add create-react-app to the project.
- Serve static assets built by create-react-app using Go.

# DONE:

- Implemented full CRUD in GraphQL.
- Implemented support for APQ running on Redis.
- Implemented aggregation pipelines for many to many by reference in MongoDB.
- Implemented marshal and unmarshal for ObjectIDs provided by `bson/primitive` package.
- Implemented custom authentication middleware.
- Added session persistence in Redis.
- Added CSRF middleware to protect the server.

# Nice To Haves:

- Add `dataloaden` to prevent N+1 problem.
- Add schema directives for authorization.
