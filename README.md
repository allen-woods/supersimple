# supersimple

A research and iteration repo for learning gqlgen. Implementing aggregation in MongoDB, go-redis, and APQ.

This was a non-project research repo for cementing fundamentals of Gqlgen, as well as exploring implementations of supporting technologies, such as Redis, Apollo Client, APQ, and MongoDB.

The repo has since evolved into a proof of concept that will serve as the basis for a project app that is currently untitled and in the concept phase.

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
