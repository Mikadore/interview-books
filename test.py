from re import A
import requests, random, string


URL = "http://0.0.0.0:8080"

def rand_str():
    return ''.join(random.choice(string.ascii_letters) for _ in range(10))

if __name__ == "__main__":
    # run twice to proove process is repeatable (i.e. no state was corrupted in the server)
    for i in range(2):
        shelves = requests.get(URL + "/shelves")
        shelves = shelves.json()
        print(f"Shelves: {shelves}")

        for shelf in shelves:
            books = requests.get(URL + f"/shelves/{shelf}")
            print(f"Shelf ({shelf}): {books.json()}")
            books = list()

            for i in range(10):
                book = {
                    "title": rand_str(),
                    "author": rand_str(),
                    "isbn": rand_str(),
                    "pages": random.randint(100, 500)
                }
                books.append(book)
                
                resp = requests.post(URL + f"/shelves/{shelf}/{book['isbn']}", json=book)
                print(f"Added book: s={resp.status_code}, resp={resp.text}: {book}")
            
            for book in books:
                resp = requests.get(URL + f"/shelves/{shelf}/{book['isbn']}")
                print(f"Fetching '{book['isbn']}': s={resp.status_code}: {resp.json()}")

            for book in books:
                print(f"Patching {book['isbn']}")
                author = book['author']
                book['author'] = book['author'][::-1]
                resp = requests.patch(URL + f"/shelves/{shelf}/{book['isbn']}", json=book)
                print(f"Patched: s={resp.status_code}: {resp.text}")

                resp = requests.get(URL + f"/shelves/{shelf}/{book['isbn']}")
                print(f"Updated: s={resp.status_code}: Old: '{author}' New: '{resp.json()['author']}'")

            for book in books:
                resp = requests.delete(URL + f"/shelves/{shelf}/{book['isbn']}")
                print(f"Deleted '{book['isbn']}': s={resp.status_code}: {resp.text}")

            print("Round trip done, shelf should be empty")
            books = requests.get(URL + f"/shelves/{shelf}")
            print(f"Shelf ({shelf}): {books.json()}")

            
        
    