# repository-pattern

Implements `Repository Pattern` in `Go`.

Normally when implementing `repository pattern` peoples tends to forgot/not considering `transaction`. Thus makes
the `service layer`/`usecase layer` to manually deal with the underlying `database driver` directly, which defeat
the purpose of using `repository pattern` in the first place.

In `java` we usually use `spring` to create `transaction` and *wire* it to `repository` for us, thus enable
us to write `repository` without thinking about `transaction`.
In `Go` there is no *`magic`* to do that, we must implement it by ourselves. This `repository` try to implement 
`repository` with `transaction` supports.

This repository uses two persistence type (`mongodb` and `postgresql`) that you can switch one to another without affecting
the logic that use the repository.

See [`main`](https://github.com/hendratommy/repository-pattern/tree/master/cmd/main.go) to try it out.
See [`tests`](https://github.com/hendratommy/repository-pattern/tree/master) for more detailed.
