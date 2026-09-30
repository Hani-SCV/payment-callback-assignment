export default () => ({
  app: {
    port: Number(process.env.PORT ?? 3000),
  },

  database: {
    host: process.env.DB_HOST ?? '127.0.0.1',
    port: Number(process.env.DB_PORT ?? 53306),
    username: process.env.MYSQL_USER ?? 'assignment',
    password: process.env.MYSQL_PASSWORD ?? 'assignment',
    database: process.env.MYSQL_DATABASE ?? 'assignment',
  },
});