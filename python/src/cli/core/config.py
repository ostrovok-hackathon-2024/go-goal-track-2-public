from pydantic import Field, ConfigDict
import yaml
from pydantic_settings import BaseSettings
from typing import List, Optional


class Settings(BaseSettings):
    MODELS_DIR: str = Field(default="../artifacts/models", alias="models_dir")
    CATEGORIES: List[str] = Field(alias="categories")

    model_config = ConfigDict(case_sensitive=False, extra="ignore")

    @classmethod
    def from_yaml(cls, yaml_file: str):
        with open(yaml_file, "r") as file:
            config = yaml.safe_load(file)
        return cls(**config)


def load_settings(config_path: Optional[str] = None) -> Settings:
    if config_path:
        return Settings.from_yaml(config_path)
    else:
        return Settings.from_yaml("config.yaml")
