# Build and deployment configuration for AWS Lambda

FROM public.ecr.aws/lambda/python:3.9

# Copy requirements and install dependencies
COPY requirements.txt ${LAMBDA_TASK_ROOT}/
RUN pip install -r requirements.txt

# Copy source code
COPY src/ ${LAMBDA_TASK_ROOT}/

# Set the CMD to your handler
CMD ["main.lambda_handler"]
