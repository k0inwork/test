import requests
import json
import os
import argparse
from bs4 import BeautifulSoup

def main():
    parser = argparse.ArgumentParser(description="Scrape legacy Django application for JSON mock data.")
    parser.add_argument("--url", default="https://c0783bc74cf80eca-77-234-200-226.serveousercontent.com", help="Base URL of the legacy app")
    parser.add_argument("--username", help="Login username", required=True)
    parser.add_argument("--password", help="Login password", required=True)
    parser.add_argument("--outdir", default="mocks", help="Output directory for saved JSON files")

    args = parser.parse_args()

    if not os.path.exists(args.outdir):
        os.makedirs(args.outdir)

    session = requests.Session()

    print(f"Connecting to {args.url}/accounts/login/ ...")
    response = session.get(f"{args.url}/accounts/login/")
    if response.status_code != 200:
        print(f"Failed to load login page: {response.status_code}")
        exit(1)

    soup = BeautifulSoup(response.text, 'html.parser')
    csrf_input = soup.find('input', {'name': 'csrfmiddlewaretoken'})
    if not csrf_input:
        print("Could not find CSRF token")
        exit(1)

    csrf_token = csrf_input['value']
    print(f"Got CSRF token: {csrf_token}")

    login_data = {
        'csrfmiddlewaretoken': csrf_token,
        'username': args.username,
        'password': args.password,
        'next': '/'
    }

    headers = {
        'Referer': f"{args.url}/accounts/login/"
    }

    print("Attempting login...")
    login_response = session.post(f"{args.url}/accounts/login/", data=login_data, headers=headers)

    if login_response.url == f"{args.url}/accounts/login/" and login_response.status_code == 200:
        print("Login failed, redirected back to login page.")
        exit(1)
    else:
        print("Login successful.")

    endpoints = {
        "products": "/products/?json=true",
        "ipmi_list": "/data/ipmi/list?json=true",
        "switch_list": "/data/switch/?json=true",
        "currentuser": "/accounts/currentuser?json=true",
        "settings": "/core/settings?json=true&json_object_list",
        "tasks": "/tasks/viewtasks/"
    }

    for name, ep in endpoints.items():
        url = f"{args.url}{ep}"
        print(f"Scraping {name} from {ep} ...")
        resp = session.get(url)
        if resp.status_code == 200 and 'application/json' in resp.headers.get('content-type', ''):
            filepath = os.path.join(args.outdir, f"{name}.json")
            try:
                data = resp.json()
                with open(filepath, "w") as f:
                    json.dump(data, f, indent=2, ensure_ascii=False)
                print(f"  -> Saved to {filepath}")
            except Exception as e:
                print(f"  -> Failed to parse JSON: {e}")
        else:
            print(f"  -> Failed to scrape {name}. Status: {resp.status_code}, Content-Type: {resp.headers.get('content-type')}")

if __name__ == "__main__":
    main()
