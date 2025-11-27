from fastapi import FastAPI, Query
from datetime import datetime, timezone
import random

app = FastAPI(title="temperature-api")


@app.get("/temperature")
def get_temperature(location: str = Query(..., description="Локация датчика")):
    """
    Эндпоинт имитации удалённого датчика.
    На каждый запрос возвращает случайную температуру.
    """
    value = round(random.uniform(-25.0, 35.0), 1)  # от -25 до +35

    return {
        "value": value,
        "unit": "C",
        "timestamp": datetime.now(timezone.utc).isoformat(),
        "location": location,
    }
