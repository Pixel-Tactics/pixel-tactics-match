import json
from websocket import create_connection

class Player:
    def __init__(self, player_id: str, player_token: str):
        self.ws = create_connection("ws://127.0.0.1:8000/ws")
        self.id = player_id
        self.token = player_token
        self.cnt = 0
    
    def _generate_id(self):
        self.cnt += 1
        return str(self.cnt)
    
    def auth(self):
        self.ws.send(json.dumps({
            "action": "AUTH",
            "identifier": self._generate_id(),
            "body": {
                "playerToken": self.token,
            },
        }))
        
    def create_session(self, opponent_id):
        self.ws.send(json.dumps({
            "action": "CREATE_SESSION",
            "identifier": self._generate_id(),
            "body": {
                "opponentId": opponent_id,
            },
        }))
        
    def prepare(self, hero_list):
        self.ws.send(json.dumps({
            "action": "PREPARE_PLAYER",
            "identifier": self._generate_id(),
            "body": {
                "chosenHeroList": hero_list,
            },
        }))
        
    def move(self):
        self.ws.send(json.dumps({
            "action": "EXECUTE_ACTION",
            "identifier": self._generate_id(),
            "body": {
                "actionName": "move",
                "actionSpecific": {
                    "heroName": "knight",
                    "directionList": ["UP"],
                }
            },
        }))
    
    def receive(self):
        return self.ws.recv()

# Player Init
player1 = Player("1", "eyJhbGciOiJIUzM4NCJ9.eyJzdWIiOiJ4eXoiLCJpYXQiOjE3MTc3NTA4OTUsImV4cCI6MTcxODExMDg5NX0.q_vubtLgMVtl8DCqjDkdOrXsbeY-wcfX_A9Eax-iO61O8TNbUyzfFzoVCHoPhHIp")
player2 = Player("2", "eyJhbGciOiJIUzM4NCJ9.eyJzdWIiOiJ4eXoyIiwiaWF0IjoxNzE3NzUxOTEzLCJleHAiOjE3MTgxMTE5MTN9.RU9z8knKK752MF36dPIhdxFVqVpqREMrKfOtnpX1avbvCpU06vU7yCpw5CI0f_ZD")

# Auth
player1.auth()
player2.auth()
print("===== Authentication =====")
print(player1.receive())
print(player2.receive())
print()

# # Create Session
# print("===== Create Session =====")
# player1.create_session(player2.id)
# print(player1.receive())
# print(player2.receive())
# player2.create_session(player1.id)
# print(player2.receive())
# print(player2.receive())
# print(player1.receive())
# print()

# # Prepare Player
# print("===== Prepare Player =====")
# player1.prepare(["knight"])
# player2.prepare(["knight"])
# print(player1.receive())
# print(player1.receive())
# print(player2.receive())
# print(player2.receive())
# print()

# # Battle
# print("===== Player 1 Turn =====")
# player1.move()
# print(player1.receive())
# print(player2.receive())
