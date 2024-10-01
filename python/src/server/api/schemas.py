from pydantic import BaseModel, Field
from typing import List, Optional
from shared.config.config import settings


class RateNameInput(BaseModel):
    rate_names: List[str] = Field(
        alias="rate_names",
        description="List of rate names to predict",
        min_items=1,
        example=["Deluxe Ocean View", "Standard Room"]
    )
    categories: Optional[List[str]] = Field(
        alias="categories",
        default=None,
        description="Categories to predict. If not provided, all available categories will be used.",
        min_items=1,
        max_items=len(settings.CATEGORIES),
        example=["capacity", "view", "bedding"],
        enum=settings.CATEGORIES
    )
