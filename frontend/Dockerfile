# syntax=docker/dockerfile:1
FROM node:22-alpine
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
EXPOSE 5173
# Dev server with host binding so the port is reachable from the host.
# (The TanStack Start production build does not emit .output/server/index.mjs
# that `npm run start` expects, so we run the Vite dev server in-container.)
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]
