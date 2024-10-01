import time
from fastapi import APIRouter, HTTPException, UploadFile, File, Form
from fastapi.responses import StreamingResponse
from .schemas import RateNameInput
from ..models.registry import model_registry
import pandas as pd
from io import StringIO

router = APIRouter()


@router.post("/predict", response_model=list[dict[str, str]])
async def predict_rate_names(input_data: RateNameInput):
    # Validate input_data.rate_names to ensure no NaN values
    cleaned_rate_names = [name if isinstance(name, str) else "" for name in input_data.rate_names]

    try:
        return model_registry.predict(cleaned_rate_names, input_data.categories)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/predict_csv")
async def predict_rate_names_csv(file: UploadFile = File(...)):
    try:
        contents = await file.read()
        df = pd.read_csv(StringIO(contents.decode("utf-8")))

        if "rate_name" not in df.columns:
            raise HTTPException(
                status_code=400, detail="CSV file must contain a 'rate_name' column"
            )

        # Replace NaN in 'rate_name' with empty string
        df['rate_name'] = df['rate_name'].fillna("")

        rate_names = df["rate_name"].tolist()

        predictions = model_registry.predict(rate_names)

        # Create a new DataFrame with predictions
        result_df = pd.DataFrame(predictions)

        # Create a StringIO object to store the CSV data
        csv_buffer = StringIO()
        result_df.to_csv(csv_buffer, index=False)
        csv_buffer.seek(0)

        # Return the CSV file as a response
        return StreamingResponse(
            iter([csv_buffer.getvalue()]),
            media_type="text/csv",
            headers={"Content-Disposition": "attachment; filename=predictions.csv"},
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
