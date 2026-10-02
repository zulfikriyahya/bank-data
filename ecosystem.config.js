module.exports = {
  apps: [
    {
      name: "core-app",
      cwd: "./backend",
      script: "./bin/server",
      interpreter: "none",
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: "300M",
      env: {
        APP_ENV: "production",
      },
      out_file: "./logs/out.log",
      error_file: "./logs/error.log",
      merge_logs: true,
      time: true,
    },
    {
      name: "scraping-import-emis",
      cwd: "./workers/scraping-import-emis",
      script: "./venv/bin/python3",
      args: "scraper.py resume",
      interpreter: "none",
      cron_restart: "0 1 * * *",
      autorestart: false,
      out_file: "./logs/out.log",
      error_file: "./logs/error.log",
      merge_logs: true,
      time: true,
    },
  ],
};
