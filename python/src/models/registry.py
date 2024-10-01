import os
import joblib
import numpy as np
from catboost import CatBoostClassifier
from sklearn.preprocessing import LabelEncoder
from ..core.config import settings
from ..core.constants import CATEGORIES

class ModelRegistry:
    def __init__(self):
        self.models = {}
        self.label_encoders = {}
        self.tfidf = None
        self.load_models()

    def load_models(self):
        self.tfidf = joblib.load(f"{settings.MODELS_DIR}/tfidf_vectorizer.joblib")

        for category in CATEGORIES:
            model_path = os.path.join(f"{settings.MODELS_DIR}/cbm", f"catboost_model_{category}.cbm")
            model = CatBoostClassifier()
            model.load_model(model_path)
            self.models[category] = model

            le = LabelEncoder()
            le.classes_ = np.load(
                f"{settings.MODELS_DIR}/label_text/label_encoder_{category}.npy", allow_pickle=True
            )
            self.label_encoders[category] = le

    def predict(self, rate_names):
        input_data = np.array(rate_names)
        input_tfidf = self.tfidf.transform(input_data)
        results = []

        for category in CATEGORIES:
            predictions = self.models[category].predict(input_tfidf)
            predictions = predictions.ravel()
            decoded_predictions = self.label_encoders[category].inverse_transform(predictions)
            results.append(decoded_predictions)

        return [
            {
                category: value.item() if isinstance(value, np.integer) else value
                for category, value in zip(CATEGORIES, row)
            }
            for row in zip(*results)
        ]

model_registry = ModelRegistry()