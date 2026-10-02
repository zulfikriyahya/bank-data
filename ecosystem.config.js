module.exports = {
  apps: [
    {
      name: "bankdata-backend",
      cwd: "./backend",
      script: "./bin/server", // hasil build go, lihat catatan di bawah
      interpreter: "none",
      env: {
        APP_ENV: "production",
        PORT: 8080,
      },
      watch: false,
      out_file: "./backend/logs/out.log",
      error_file: "./backend/logs/error.log",
      merge_logs: true,
    },
    {
      name: "bankdata-frontend",
      cwd: "./frontend",
      script: "npm",
      args: "run start", // pastikan sudah `npm run build` dulu
      env: {
        NODE_ENV: "production",
        PORT: 3000,
      },
      watch: false,
      out_file: "./frontend/logs/out.log",
      error_file: "./frontend/logs/error.log",
      merge_logs: true,
    },
  ],
};
