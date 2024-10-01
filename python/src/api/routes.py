from fastapi import APIRouter, HTTPException
from .schemas import RateNameInput
from ..models.registry import model_registry

router = APIRouter()


@router.post("/predict", response_model=list[dict[str, str]])
async def predict_rate_names(input_data: RateNameInput):
    try:
        return model_registry.predict(input_data.rate_names, input_data.categories)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
