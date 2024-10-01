from functools import lru_cache
import os
import joblib
import numpy as np
from catboost import CatBoostClassifier
from sklearn.preprocessing import LabelEncoder
from concurrent.futures import ThreadPoolExecutor, as_completed


class ModelRegistry:
    def __init__(self, models_dir: str, categories: list[str], max_workers: int = None):
        self.models_dir = models_dir
        self.categories = categories
        self.models = {}
        self.label_encoders = {}
        self.tfidf = None
        self.max_workers = max_workers or os.cpu_count()
        self.load_models()

    @lru_cache(maxsize=None)
    def _load_single_model(self, category: str):
        model_path = os.path.join(
            self.models_dir, "cbm", f"catboost_model_{category}.cbm"
        )
        model = CatBoostClassifier()
        model.load_model(model_path)

        le_path = os.path.join(
            self.models_dir, "label_text", f"label_encoder_{category}.npy"
        )
        le = LabelEncoder()
        le.classes_ = np.load(le_path, allow_pickle=True)

        return category, model, le

    def load_models(self):
        self.tfidf = joblib.load(
            os.path.join(self.models_dir, "tfidf_vectorizer.joblib")
        )

        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            futures = {
                executor.submit(self._load_single_model, cat): cat
                for cat in self.categories
            }
            for future in as_completed(futures):
                category, model, le = future.result()
                self.models[category] = model
                self.label_encoders[category] = le

    def predict_category(self, category: str, input_tfidf: np.ndarray):
        model = self.models[category]
        le = self.label_encoders[category]
        predictions = model.predict(input_tfidf)
        predictions = predictions.ravel()
        decoded = le.inverse_transform(predictions)
        return category, decoded

    def predict(self, rate_names: list[str], categories: list[str] = None):
        if categories is None or len(categories) == 0:
            categories = self.categories
        else:
            categories = [cat for cat in categories if cat in self.categories]

        input_tfidf = self.tfidf.transform(rate_names)
        results = {category: None for category in categories}

        with ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            futures = {
                executor.submit(self.predict_category, cat, input_tfidf): cat
                for cat in categories
            }
            for future in as_completed(futures):
                category, decoded_predictions = future.result()
                results[category] = decoded_predictions

        return [
            {
                "rate_name": rate_name,
                **{
                    category: value.item() if isinstance(value, np.integer) else value
                    for category, value in zip(categories, row)
                },
            }
            for rate_name, row in zip(
                rate_names, zip(*(results[cat] for cat in categories))
            )
        ]
