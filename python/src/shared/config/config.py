from pydantic import Field
import yaml
from pydantic_settings import BaseSettings
from typing import List


class Settings(BaseSettings):
    MODELS_DIR: str = Field(default="../artifacts/models", alias="models_dir")
    CATEGORIES: List[str] = Field(alias="categories")

    class Config:
        case_sensitive = False
        extra = "ignore"

    @classmethod
    def from_yaml(cls, yaml_file: str):
        with open(yaml_file, "r") as file:
            config = yaml.safe_load(file)
        return cls(**config)


def load_settings(config_path: str = "config.yaml") -> Settings:
    return Settings.from_yaml(config_path)
