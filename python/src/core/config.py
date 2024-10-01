from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PROJECT_NAME: str = "Rate Name Classifier"
    VERSION: str = "0.1.0"
    MODELS_DIR: str = "../artifacts/models"


settings = Settings()
