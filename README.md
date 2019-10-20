# supersimple

A research and iteration repo for learning gqlgen. Implementing aggregation in MongoDB, go-redis, and APQ.

This is a non-project research repo for cementing fundamentals of Gqlgen, as well as exploring implementations of supporting technologies, such as Redis, Apollo Client, APQ, and MongoDB.

# TODO:

- Implement Book and Author models.
- Use aggregation to join Authors and Books:
  - Books contain an array of Author IDs.
  - Authors can match their ID with Books.
