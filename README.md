## Development

```bash
go run main.go
```

For hot reloading, use [`air`](https://github.com/cosmtrek/air):
```bash
air
```

### Environment Variables
Generate a 32-byte key for `SESSION_SECRET` and make sure it's available in your environment.
An `.env.example` file is provided in case you want to cp into `.env` to use in development.

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