from fastapi import APIRouter, HTTPException
from .schemas import RateNameInput, RateNamePrediction
from ..models.registry import model_registry

router = APIRouter()


@router.post("/predict", response_model=list[RateNamePrediction])
async def predict_rate_names(input_data: RateNameInput):
    try:
        predictions = model_registry.predict(input_data.rate_names)
        return [
            RateNamePrediction(
                class_=pred["class"],
                capacity=pred["capacity"],
                quality=pred["quality"],
                view=pred["view"],
                bedding=pred["bedding"],
                balcony=pred["balcony"],
                bedrooms=pred["bedrooms"],
                club=pred["club"],
                floor=pred["floor"],
                bathroom=pred["bathroom"],
            )
            for pred in predictions
        ]
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
