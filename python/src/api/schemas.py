from pydantic import BaseModel
from typing import List


class RateNameInput(BaseModel):
    rate_names: List[str]
