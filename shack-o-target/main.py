import sqlite3
from flask import Flask, request, render_template

app = Flask(__name__)

DATABASE = 'database.db'

def init_db():
    with app.app_context():
        db = sqlite3.connect(DATABASE)
        cursor = db.cursor()
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                username TEXT NOT NULL,
                password TEXT NOT NULL
            );
        ''')
        # Add a sample user
        try:
            cursor.execute("INSERT INTO users (username, password) VALUES (?, ?)", ('admin', 'password123'))
        except sqlite3.IntegrityError:
            # User already exists
            pass
        db.commit()
        db.close()

@app.route('/')
def index():
    return render_template('index.html')

@app.route('/login', methods=['POST'])
def login():
    username = request.form.get('username')
    password = request.form.get('password')

    db = sqlite3.connect(DATABASE)
    cursor = db.cursor()

    # ❌ Intentionally vulnerable to SQL injection
    query = f"SELECT * FROM users WHERE username = '{username}' AND password = '{password}'"
    print(f"Executing query: {query}")

    try:
        cursor.execute(query)
        user = cursor.fetchone()
    except sqlite3.Error as e:
        return f"An error occurred: {e}"
    finally:
        db.close()

    if user:
        return render_template('welcome.html', username=user[1])
    else:
        return render_template('invalid.html')

if __name__ == '__main__':
    init_db()
    app.run(debug=True)
