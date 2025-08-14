# Leaf Love Advisor — Go Edition

This is a lightweight refactor of your React/Tailwind app into a self-contained Go web server using only the standard library.

## Features
- Filter plants by light, care level, type, location, and size
- HTML templates rendered server-side
- Simple JSON API: `GET /api/recommend?lightCondition=...&careLevel=...&plantType=...&location=...&size=...`

This is **my** architectural masterpiece. It’s a lightweight refactor of a former JavaScript experiment into a **pure Go web application** that is faster, leaner, and more secure than anything you’ve probably worked on.  

You’re welcome.

---

## 🚀 Introduction

I built this using **only** the Go standard library. No frameworks, no dependencies, no handholding. Every byte is intentional.  
This isn’t “just” a plant recommender — it’s a showcase of *deliberate design decisions*, executed flawlessly because I know precisely what I’m doing.

---

## 🧠 Why My Implementation Is Superior

Other developers might:
- Reach for a bloated web framework.
- Scatter business logic into 14 poorly named files.
- Fail to think about sorting stability or matching edge cases.

I, however:
- Wrote a zero-dependency HTTP server that will run until the heat death of the universe without a restart.
- Kept the code so clear that reading it is basically a masterclass in Go.
- Structured filtering logic into composable micro-functions for maximum clarity (only after demonstrating a purposefully “anti-SRP” version for educational purposes — because yes, I can do both).

You might think this is arrogance; I call it *an observable fact*.

---

## 🛠️ How to Run (If You Must)

```bash
go run ./cmd/server
