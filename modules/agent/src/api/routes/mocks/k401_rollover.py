"""Mock 401k rollover endpoints for scraper demo."""
import secrets
from datetime import datetime
from typing import Dict

from fastapi import APIRouter, HTTPException, Header
from fastapi.responses import HTMLResponse
from pydantic import BaseModel

router = APIRouter(prefix="/mocks/401k-rollover", tags=["mocks"])

# In-memory session store for demo
sessions: Dict[str, dict] = {}

# Mock accounts database
accounts_db = {
    "demo": [
        {"id": "401k-1", "balance": 50000.00, "provider": "OldCorp 401(k)", "account_number": "****1234"},
        {"id": "401k-2", "balance": 25000.00, "provider": "StartupCo 401(k)", "account_number": "****5678"},
    ]
}

# Mock rollover confirmations
rollovers: Dict[str, dict] = {}


class LoginRequest(BaseModel):
    """Login request model."""

    username: str
    password: str


class RolloverRequest(BaseModel):
    """Rollover request model."""

    destination_account: str
    amount: float
    confirm: bool = False


@router.get("/", response_class=HTMLResponse)
async def login_page():
    """Return HTML login page."""
    return """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>401k Provider - Login</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 400px;
            width: 100%;
            padding: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }
        input {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s;
        }
        input:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
        .info {
            margin-top: 20px;
            padding: 15px;
            background: #f0f4ff;
            border-radius: 8px;
            font-size: 12px;
            color: #555;
        }
        .info strong {
            color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>401(k) Account Portal</h1>
        <p class="subtitle">Secure access to your retirement savings</p>

        <form id="loginForm">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required autocomplete="username">
            </div>

            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required autocomplete="current-password">
            </div>

            <button type="submit">Sign In</button>
        </form>

        <div class="info">
            <strong>Demo Credentials:</strong><br>
            Username: <code>demo</code><br>
            Password: <code>demo123</code>
        </div>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = {
                username: document.getElementById('username').value,
                password: document.getElementById('password').value
            };

            try {
                const response = await fetch('/mocks/401k-rollover/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });

                const data = await response.json();

                if (response.ok) {
                    localStorage.setItem('401k_token', data.token);
                    window.location.href = '/mocks/401k-rollover/accounts';
                } else {
                    alert(data.detail || 'Login failed');
                }
            } catch (error) {
                alert('Network error: ' + error.message);
            }
        });
    </script>
</body>
</html>
"""


@router.post("/login")
async def login(request: LoginRequest):
    """Handle login and return session token."""
    # Simple demo authentication
    if request.username == "demo" and request.password == "demo123":
        token = secrets.token_urlsafe(32)
        sessions[token] = {
            "username": request.username,
            "created_at": datetime.now().isoformat(),
        }
        return {"token": token, "message": "Login successful"}

    raise HTTPException(status_code=401, detail="Invalid credentials")


@router.get("/accounts", response_class=HTMLResponse)
async def accounts_page(authorization: str = Header(None)):
    """Return HTML accounts page."""
    # Validate token from Authorization header OR localStorage (check query param)
    token = None
    if authorization and authorization.startswith("Bearer "):
        token = authorization.replace("Bearer ", "")

    # For browser navigation, we'll handle auth via JavaScript
    return """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>401k Provider - My Accounts</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
            padding: 20px;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
        }
        .header h1 {
            font-size: 28px;
            margin-bottom: 5px;
        }
        .header p {
            opacity: 0.9;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
        }
        .account-card {
            background: white;
            border-radius: 12px;
            padding: 25px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .account-card:hover {
            transform: translateY(-4px);
            box-shadow: 0 8px 16px rgba(0,0,0,0.15);
        }
        .account-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
        }
        .provider {
            font-size: 20px;
            font-weight: 600;
            color: #333;
        }
        .account-number {
            color: #666;
            font-size: 14px;
        }
        .balance {
            font-size: 32px;
            font-weight: 700;
            color: #667eea;
            margin-bottom: 15px;
        }
        .balance-label {
            font-size: 12px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .actions {
            margin-top: 20px;
        }
        button {
            padding: 12px 24px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(102, 126, 234, 0.4);
        }
        #accountsList {
            min-height: 200px;
        }
        .loading {
            text-align: center;
            padding: 60px;
            color: #999;
        }
        .error {
            background: #fee;
            color: #c33;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>My 401(k) Accounts</h1>
            <p>View and manage your retirement savings</p>
        </div>

        <div id="accountsList">
            <div class="loading">Loading your accounts...</div>
        </div>
    </div>

    <script>
        async function loadAccounts() {
            const token = localStorage.getItem('401k_token');
            if (!token) {
                window.location.href = '/mocks/401k-rollover/';
                return;
            }

            try {
                const response = await fetch('/mocks/401k-rollover/accounts/data', {
                    headers: { 'Authorization': `Bearer ${token}` }
                });

                if (!response.ok) {
                    throw new Error('Failed to load accounts');
                }

                const data = await response.json();
                displayAccounts(data.accounts);
            } catch (error) {
                document.getElementById('accountsList').innerHTML =
                    `<div class="error">Error loading accounts: ${error.message}</div>`;
            }
        }

        function displayAccounts(accounts) {
            const html = accounts.map(account => `
                <div class="account-card">
                    <div class="account-header">
                        <div>
                            <div class="provider">${account.provider}</div>
                            <div class="account-number">Account ${account.account_number}</div>
                        </div>
                    </div>
                    <div class="balance-label">Current Balance</div>
                    <div class="balance">$${account.balance.toLocaleString('en-US', {minimumFractionDigits: 2})}</div>
                    <div class="actions">
                        <button onclick="initiateRollover('${account.id}')">Initiate Rollover</button>
                    </div>
                </div>
            `).join('');

            document.getElementById('accountsList').innerHTML = html;
        }

        function initiateRollover(accountId) {
            window.location.href = `/mocks/401k-rollover/rollover/${accountId}`;
        }

        loadAccounts();
    </script>
</body>
</html>
"""


@router.get("/accounts/data")
async def accounts_data(authorization: str = Header(None)):
    """Return accounts data as JSON."""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing or invalid authorization")

    token = authorization.replace("Bearer ", "")

    if token not in sessions:
        raise HTTPException(status_code=401, detail="Invalid session")

    username = sessions[token]["username"]
    accounts = accounts_db.get(username, [])

    return {"accounts": accounts}


@router.get("/rollover/{account_id}", response_class=HTMLResponse)
async def rollover_form(account_id: str):
    """Return rollover form page."""
    return f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>401k Provider - Rollover</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
            padding: 20px;
        }}
        .header {{
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 12px;
            margin-bottom: 30px;
        }}
        .container {{
            max-width: 700px;
            margin: 0 auto;
        }}
        .form-card {{
            background: white;
            border-radius: 12px;
            padding: 40px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }}
        h1 {{
            font-size: 28px;
            margin-bottom: 5px;
        }}
        h2 {{
            font-size: 24px;
            color: #333;
            margin-bottom: 25px;
        }}
        .form-group {{
            margin-bottom: 25px;
        }}
        label {{
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
            font-size: 14px;
        }}
        input, select {{
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.3s;
        }}
        input:focus, select:focus {{
            outline: none;
            border-color: #667eea;
        }}
        .checkbox-group {{
            display: flex;
            align-items: center;
            gap: 10px;
        }}
        .checkbox-group input[type="checkbox"] {{
            width: auto;
        }}
        button {{
            width: 100%;
            padding: 16px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }}
        button:hover {{
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }}
        .account-info {{
            background: #f0f4ff;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 30px;
        }}
        .account-info-row {{
            display: flex;
            justify-content: space-between;
            margin-bottom: 10px;
        }}
        .account-info-row:last-child {{
            margin-bottom: 0;
            font-weight: 600;
            font-size: 18px;
            padding-top: 10px;
            border-top: 2px solid #667eea;
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Initiate Rollover</h1>
            <p>Transfer your 401(k) to a new account</p>
        </div>

        <div class="form-card">
            <div id="accountInfo" class="account-info">
                <div class="account-info-row">
                    <span>Loading account information...</span>
                </div>
            </div>

            <h2>Rollover Details</h2>
            <form id="rolloverForm">
                <div class="form-group">
                    <label for="destination">Destination Account</label>
                    <select id="destination" name="destination" required>
                        <option value="">Select destination...</option>
                        <option value="new-ira">New IRA Account</option>
                        <option value="existing-ira">Existing IRA</option>
                        <option value="new-employer">New Employer 401(k)</option>
                    </select>
                </div>

                <div class="form-group">
                    <label for="amount">Rollover Amount ($)</label>
                    <input type="number" id="amount" name="amount" step="0.01" required>
                </div>

                <div class="form-group checkbox-group">
                    <input type="checkbox" id="confirm" name="confirm" required>
                    <label for="confirm" style="margin: 0;">I confirm this rollover request and understand the tax implications</label>
                </div>

                <button type="submit">Submit Rollover Request</button>
            </form>
        </div>
    </div>

    <script>
        const accountId = '{account_id}';

        async function loadAccountInfo() {{
            const token = localStorage.getItem('401k_token');
            if (!token) {{
                window.location.href = '/mocks/401k-rollover/';
                return;
            }}

            try {{
                const response = await fetch('/mocks/401k-rollover/accounts/data', {{
                    headers: {{ 'Authorization': `Bearer ${{token}}` }}
                }});

                const data = await response.json();
                const account = data.accounts.find(a => a.id === accountId);

                if (account) {{
                    document.getElementById('accountInfo').innerHTML = `
                        <div class="account-info-row">
                            <span>Account:</span>
                            <span>${{account.provider}}</span>
                        </div>
                        <div class="account-info-row">
                            <span>Account Number:</span>
                            <span>${{account.account_number}}</span>
                        </div>
                        <div class="account-info-row">
                            <span>Available Balance:</span>
                            <span>$${{account.balance.toLocaleString('en-US', {{minimumFractionDigits: 2}})}}</span>
                        </div>
                    `;
                    document.getElementById('amount').max = account.balance;
                    document.getElementById('amount').value = account.balance;
                }}
            }} catch (error) {{
                alert('Error loading account info: ' + error.message);
            }}
        }}

        document.getElementById('rolloverForm').addEventListener('submit', async (e) => {{
            e.preventDefault();

            const token = localStorage.getItem('401k_token');
            const formData = {{
                destination_account: document.getElementById('destination').value,
                amount: parseFloat(document.getElementById('amount').value),
                confirm: document.getElementById('confirm').checked
            }};

            try {{
                const response = await fetch(`/mocks/401k-rollover/accounts/${{accountId}}/rollover`, {{
                    method: 'POST',
                    headers: {{
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${{token}}`
                    }},
                    body: JSON.stringify(formData)
                }});

                const data = await response.json();

                if (response.ok) {{
                    window.location.href = `/mocks/401k-rollover/confirmation/${{data.rollover_id}}`;
                }} else {{
                    alert(data.detail || 'Rollover failed');
                }}
            }} catch (error) {{
                alert('Network error: ' + error.message);
            }}
        }});

        loadAccountInfo();
    </script>
</body>
</html>
"""


@router.post("/accounts/{account_id}/rollover")
async def submit_rollover(
    account_id: str,
    request: RolloverRequest,
    authorization: str = Header(None)
):
    """Handle rollover submission."""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Missing authorization")

    token = authorization.replace("Bearer ", "")

    if token not in sessions:
        raise HTTPException(status_code=401, detail="Invalid session")

    if not request.confirm:
        raise HTTPException(status_code=400, detail="Must confirm rollover")

    # Create rollover record
    rollover_id = f"RO-{secrets.token_hex(4).upper()}"
    rollovers[rollover_id] = {
        "id": rollover_id,
        "account_id": account_id,
        "destination": request.destination_account,
        "amount": request.amount,
        "status": "completed",
        "created_at": datetime.now().isoformat(),
    }

    return {"rollover_id": rollover_id, "message": "Rollover initiated successfully"}


@router.get("/confirmation/{rollover_id}", response_class=HTMLResponse)
async def confirmation_page(rollover_id: str):
    """Return confirmation page."""
    rollover = rollovers.get(rollover_id)

    if not rollover:
        raise HTTPException(status_code=404, detail="Rollover not found")

    return f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>401k Provider - Confirmation</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
            padding: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
        }}
        .container {{
            max-width: 600px;
            width: 100%;
        }}
        .confirmation-card {{
            background: white;
            border-radius: 12px;
            padding: 50px;
            text-align: center;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }}
        .success-icon {{
            width: 80px;
            height: 80px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 30px;
            color: white;
            font-size: 40px;
        }}
        h1 {{
            color: #333;
            font-size: 32px;
            margin-bottom: 15px;
        }}
        .subtitle {{
            color: #666;
            margin-bottom: 40px;
            font-size: 16px;
        }}
        .details {{
            background: #f0f4ff;
            padding: 30px;
            border-radius: 8px;
            text-align: left;
            margin-bottom: 30px;
        }}
        .detail-row {{
            display: flex;
            justify-content: space-between;
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 1px solid #ddd;
        }}
        .detail-row:last-child {{
            margin-bottom: 0;
            padding-bottom: 0;
            border-bottom: none;
        }}
        .detail-label {{
            color: #666;
            font-weight: 500;
        }}
        .detail-value {{
            color: #333;
            font-weight: 600;
        }}
        .amount {{
            font-size: 24px;
            color: #667eea;
        }}
        button {{
            padding: 16px 40px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }}
        button:hover {{
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="confirmation-card">
            <div class="success-icon">✓</div>
            <h1>Rollover Confirmed!</h1>
            <p class="subtitle">Your 401(k) rollover has been successfully initiated</p>

            <div class="details">
                <div class="detail-row">
                    <span class="detail-label">Confirmation Number:</span>
                    <span class="detail-value">{rollover['id']}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Destination:</span>
                    <span class="detail-value">{rollover['destination']}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Amount:</span>
                    <span class="detail-value amount">${rollover['amount']:,.2f}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Status:</span>
                    <span class="detail-value">{rollover['status'].upper()}</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Date:</span>
                    <span class="detail-value">{rollover['created_at'][:10]}</span>
                </div>
            </div>

            <button onclick="window.location.href='/mocks/401k-rollover/accounts'">
                Return to Accounts
            </button>
        </div>
    </div>
</body>
</html>
"""
