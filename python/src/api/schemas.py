from pydantic import BaseModel
from typing import List


class RateNameInput(BaseModel):
    rate_names: List[str]


class RateNamePrediction(BaseModel):
    class_: str
    capacity: str
    quality: str
    view: str
    bedding: str
    balcony: str
    bedrooms: str
    club: str
    floor: str
    bathroom: str


    class Config:
        from_attributes = True
        populate_by_name = True