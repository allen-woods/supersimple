# supersimple

A research and iteration repo for learning gqlgen. Implementing aggregation in MongoDB, go-redis, and APQ.

This is a non-project research repo for cementing fundamentals of Gqlgen, as well as exploring implementations of supporting technologies, such as Redis, Apollo Client, APQ, and MongoDB.

# TODO:

- Add Dataloaden to prevent N+1 problem.
- Add Indexes to the documents in the database to prevent worsening performance as data store grows.

# DONE:

- Implemented full CRUD in GraphQL.
- Implemented support for APQ running on Redis.
- Implemented aggregation pipelines for many to many by reference in MongoDB.
- Implemented marshal and unmarshal for ObjectIDs provided by `bson/primitive` package.
- Implemented Gorilla Sessions middleware.
- Added Redisstore as persisted session store.
- Added CSRF middleware to protect the server.
