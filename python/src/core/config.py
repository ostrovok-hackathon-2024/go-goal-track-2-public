from pydantic import Field
import yaml
from pydantic_settings import BaseSettings
from typing import List


class Settings(BaseSettings):
    PROJECT_NAME: str = Field(default="Rate Name Classifier", alias="project_name")
    VERSION: str = Field(default="0.1.0", alias="version")
    MODELS_DIR: str = Field(default="../artifacts/models", alias="models_dir")
    CATEGORIES: List[str] = Field(alias="categories")
    PORT: int = Field(default=8000, alias="port")

    class Config:
        case_sensitive = False
        extra = "ignore"

    @classmethod
    def from_yaml(cls, yaml_file: str):
        with open(yaml_file, "r") as file:
            config = yaml.safe_load(file)
        return cls(**config)


settings = Settings.from_yaml("config.yaml")
