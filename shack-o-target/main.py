from flask import Flask, request, render_template_string

app = Flask(__name__)

@app.route('/')
def index():
    return '''
    <h2>Login Form</h2>
    <form action="/login" method="post">
        Username: <input name="username"><br>
        Password: <input name="password" type="password"><br>
        <button type="submit">Login</button>
    </form>
    '''

@app.route('/login', methods=['POST'])
def login():
    username = request.form.get('username')
    password = request.form.get('password')

    # ❌ Intentionally vulnerable (string concatenation)
    query = f"SELECT * FROM users WHERE username = '{username}' AND password = '{password}'"
    print("Executing:", query)

    # Fake response — we’re not actually querying a DB
    if "admin" in username and "password" in password:
        return "Welcome admin!"
    else:
        return "Invalid credentials!"

if __name__ == '__main__':
    app.run(debug=True)
