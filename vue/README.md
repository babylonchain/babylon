## Project setup

### node setup
1. In your local node's ```babylon/config/config.toml``` file, under the "RPC Server Configuration Options" section, set ```cors_allowed_origins = ["*"]```
2. In the same file under the "Instrumentation Configuration Options" seciton, set ```prometheus = true``` and ```max_open_connections = 0```
3. In the ```babylon/config/app.toml``` file, under the ```API Configuration``` section, set ```enabled-unsafe-cors = true```

### Web app setup

```
npm install
```

### Compiles and reloads the app on save for development

```
npm run dev
```

### Compiles and minifies for production

```
npm run build
```
