# githooks/

可选：本地开发 hooks（CI 里有重跑，不装也不会影响 merge）。

安装：
```bash
cd repo-root
git config core.hooksPath githooks
```

钩子：
- `pre-commit`：跑 gofmt + go vet + 禁止硬编码 sk-xxx + 禁止 InsecureSkipVerify。
- `commit-msg`：标题必须 `feat/fix/chore/docs/build/test/refactor/style/ci:` 开头，
  并强制 commit author/committer 必须是 `<johnmelodymel@qq.com>`（弟弟的真实邮箱）。
- `pre-push`：可选跑 `make test -short`，防止把红的推到远端。

hooks 里有任何 false positive，先改 hook，不要 `--no-verify` 绕过。绕过一次就告诉二姐哦 😊
