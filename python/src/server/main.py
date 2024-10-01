import os
from fastapi import FastAPI
import uvicorn
from .api.routes import router

app = FastAPI(title="Tagger API", version="0.1.0")


app.include_router(router)


def main():
    port = os.getenv("PORT", 8080)
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()
