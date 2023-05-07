import requests


def test_endpoint(url: str):
    response = requests.get(url)

    if response.status_code == 200:
        try:
            json_response = response.json()
            print(json_response)
        except requests.exceptions.JSONDecodeError as e:
            print(f"JSON decoding error: {e}")
            print(f"Response text: {response.text}")
    else:
        print(f"Request failed with status code {response.status_code}")


if __name__ == "__main__":
    base_url = 'http://localhost:8080'
    #test_endpoint(url=f'{base_url}/order/create/neworder123')
    #test_endpoint(url=f'{base_url}/payment/pay/user123/order123/amount123')
    test_endpoint(url=f'{base_url}/stock/find/item123/')
