# 👹 !!СОБЕРИ МЕНЯ, ЕСЛИ СМОЖЕШЬ!!! 👹


# Use the official PyPy image as the base
FROM pypy:3.10-7.3.17-bookworm

# Set the working directory in the container
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    gfortran \
    libopenblas-dev \
    && rm -rf /var/lib/apt/lists/*

# Upgrade pip and install wheel
RUN pip install --upgrade pip wheel

# Install Python packages, preferring binary wheels
RUN pip install --no-cache-dir --prefer-binary \
    scipy \ 
    numpy \
    cython \
    catboost \
    optuna \
    dill \
    psutil \
    scikit-learn

# Copy your application files
COPY . .

# Run the main.py script using PyPy
CMD ["pypy3", "main.py"]