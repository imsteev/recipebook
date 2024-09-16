## Development

```bash
go run main.go
```

For hot reloading, use [`air`](https://github.com/cosmtrek/air):
```bash
air
```

### TailwindCSS
This project uses [TailwindCSS](https://tailwindcss.com/) for styling. If you
have Node installed, you can use Bun or Node to watch for file changes in order
to rebuild the CSS from source files.
```
bun install
bun run watch
```

## Deployment
- [ ] build frontend assets