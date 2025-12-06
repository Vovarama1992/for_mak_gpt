from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from docx import Document
from PIL import Image, ImageDraw
import io, base64
import uvicorn

app = FastAPI()

@app.post("/convert")
async def convert(request: Request):
    raw = await request.body()

    doc = Document(io.BytesIO(raw))

    pages = []

    for i, para in enumerate(doc.paragraphs, start=1):
        img = Image.new("RGB", (1600, 2200), "white")
        d = ImageDraw.Draw(img)
        d.text((50, 50), para.text, fill="black")

        buf = io.BytesIO()
        img.save(buf, format="JPEG")
        b64 = base64.b64encode(buf.getvalue()).decode()

        pages.append({
            "file_name": f"page-{i}.jpg",
            "mime": "image/jpeg",
            "base64": b64,
        })

    return JSONResponse({"pages": pages})

# ← ВОТ ЭТО ТЫ ЗАБЫЛ
if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)