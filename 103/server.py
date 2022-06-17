from concurrent import futures
import logging

import grpc
import hello_pb2
import hello_pb2_grpc


class Greeter(hello_pb2_grpc.HelloServiceServicer):
    def SayHello(self, request, context):
        print("Say hello function called...")
        return hello_pb2.HelloResponse(message=f"Hello {request.name}")


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
    hello_pb2_grpc.add_HelloServiceServicer_to_server(Greeter(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    logging.basicConfig()
    serve()
