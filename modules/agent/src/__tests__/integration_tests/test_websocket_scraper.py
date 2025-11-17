"""Test scraper via WebSocket end-to-end."""
import asyncio
import json
import websockets


async def test_scraper_websocket():
    """Test scraper integration via WebSocket."""
    uri = "ws://localhost:8000/ws"

    print("=" * 60)
    print("Testing Scraper via WebSocket")
    print("=" * 60)

    async with websockets.connect(uri) as websocket:
        # Send a scraping request
        request = {
            "user_id": "test-user",
            "message": "Please scrape the mock 401k site at http://localhost:8000/mocks/401k-rollover/ and tell me what you find"
        }

        print(f"\nSending request: {request['message'][:80]}...")
        await websocket.send(json.dumps(request))

        print("\nReceiving responses:")
        print("-" * 60)

        response_count = 0
        while True:
            try:
                message = await asyncio.wait_for(websocket.recv(), timeout=60.0)
                data = json.dumps(message)

                response_count += 1
                print(f"\n[Response {response_count}]")
                print(message[:200] if len(message) > 200 else message)

                # Check if this looks like a final response
                if "✓" in message or "complete" in message.lower() or response_count > 10:
                    print("\n" + "-" * 60)
                    print(f"Received {response_count} responses")
                    break

            except asyncio.TimeoutError:
                print("\nTimeout waiting for response")
                break

    print("\n" + "=" * 60)
    print("WebSocket Test Complete!")
    print("=" * 60)


if __name__ == "__main__":
    asyncio.run(test_scraper_websocket())
