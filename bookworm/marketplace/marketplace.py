import os

from flask import Flask, render_template
import grpc

from recommendations_pb2 import BookCategory, RecommendationRequest
from recommendations_pb2_grpc import RecommendationsStub

app = Flask(__name__)

recommendations_addr = os.getenv("RECOMMENDATIONS_ADDR", "localhost:50051")
recommendations_channel = grpc.insecure_channel(recommendations_addr)
recommendations_client = RecommendationsStub(recommendations_channel)


@app.route("/")
def render_homepage():
    recommendations_request = RecommendationRequest(user_id=1, category=BookCategory.SELF_HELP, max_results=3)
    recommendations_response = recommendations_client.Recommend(recommendations_request)
    return render_template(
        "home.html",
        recommendations=recommendations_response.recommendations
    )

