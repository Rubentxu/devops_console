# DevOps Console

Este proyecto consiste en una aplicación de consola DevOps con un backend en FastAPI y un frontend en React.

## Estructura del Proyecto

```
devops_console/
├── backend/
└── frontend/
```

## Requisitos Previos

- Python 3.8+
- Poetry (para gestión de dependencias de Python)
- Node.js 14+
- Bun (opcional, pero recomendado para el frontend)

## Configuración

### Backend

1. Navega al directorio del backend:
   ```
   cd backend
   ```

2. Instala Poetry si aún no lo tienes:
   ```
   curl -sSL https://install.python-poetry.org | python3 -
   ```

3. Instala las dependencias del proyecto usando Poetry:
   ```
   poetry install
   ```

4. Activa el entorno virtual creado por Poetry:
   ```
   poetry shell
   ```

5. Crea un archivo `.env` en el directorio `backend/` con el siguiente contenido:
   ```
   API_PORT=8000
   ```

### Frontend

1. Navega al directorio del frontend:
   ```
   cd frontend
   ```

2. Instala las dependencias:
   ```
   npm install
   # O si estás usando Bun:
   bun install
   ```

3. Crea un archivo `.env` en el directorio `frontend/` con el siguiente contenido:
   ```
   VITE_API_URL=http://localhost:8000
   VITE_APP_PORT=3000
   VITE_APP_NAME="DevOps Console"
   ```

## Ejecución

### Backend

1. Asegúrate de estar en el directorio `backend/` y que el entorno virtual de Poetry esté activado.
2. Ejecuta el servidor:
   ```
   poetry run python main.py
   ```
   El servidor estará disponible en `http://localhost:8000`.

### Frontend

1. Asegúrate de estar en el directorio `frontend/`.
2. Ejecuta el servidor de desarrollo:
   ```
   npm run dev
   # O si estás usando Bun:
   bun run dev
   ```
   La aplicación estará disponible en `http://localhost:3000`.

## Entornos

### Desarrollo

- Usa los archivos `.env` como se describió anteriormente.
- Asegúrate de que `VITE_API_URL` en el frontend apunte a tu servidor de backend local.

### Producción

- En el backend, configura las variables de entorno en tu servidor de producción.
- Para el frontend, crea un archivo `.env.production` con la URL de tu API de producción:
  ```
  VITE_API_URL=https://api.tudominio.com
  ```
- Construye el frontend para producción:
  ```
  npm run build
  # O si estás usando Bun:
  bun run build
  ```

## Contribuir

1. Haz un fork del repositorio
2. Crea una nueva rama (`git checkout -b feature/AmazingFeature`)
3. Haz commit de tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Haz push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## Notas adicionales

- Asegúrate de no subir los archivos `.env` al repositorio. Están incluidos en `.gitignore`.
- Para el backend, usamos Poetry para manejar las dependencias y el entorno virtual.
- Para el frontend, puedes usar npm o Bun, dependiendo de tu preferencia.
