# Fullstack Starter

初始化完成的全栈项目，技术栈如下：

- 前端：TypeScript + React + Ant Design + TailwindCSS + Vite
- 后端：Golang + Gin + Gorm + MySQL
- 容器编排：Docker Compose

## 目录结构

```text
.
├── frontend/          # React 前端
├── backend/           # Go 后端
└── docker-compose.yml # 一键启动前后端 + MySQL
```

## 本地开发

### 1) 启动 MySQL（可选）

```bash
docker compose up -d mysql
```

### 2) 启动后端

```bash
cd backend
cp .env.example .env
go run .
```

后端默认地址：`http://localhost:8080`

### 3) 启动前端

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

前端默认地址：`http://localhost:5173`

## Docker Compose 一键启动

```bash
docker compose up -d --build
```

服务地址：

- 前端：http://localhost:3000
- 后端：http://localhost:8080/api/health
- MySQL：localhost:3306

停止服务：

```bash
docker compose down
```
