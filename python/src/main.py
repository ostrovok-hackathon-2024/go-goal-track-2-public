from fastapi import FastAPI
from .api.routes import router
from .core.config import settings
from .core.logging import setup_logging

app = FastAPI(title=settings.PROJECT_NAME, version=settings.VERSION)

setup_logging()

app.include_router(router)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)