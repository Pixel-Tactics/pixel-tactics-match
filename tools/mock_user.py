from flask import Flask, request

app = Flask(__name__)

@app.route('/auth/current', methods=['GET'])
def auth():
    username = request.headers.get("Authorization", "")[7:]
    return {
        "username": username
    }

if __name__ == '__main__':
    app.run(debug=True)
