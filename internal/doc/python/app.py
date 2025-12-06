from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from docx import Document
import io
import uvicorn

app = FastAPI()

@app.post("/convert")
async def convert(request: Request):
    raw = await request.body()

    # DOCX → TEXT
    try:
        doc = Document(io.BytesIO(raw))
    except Exception as e:
        return JSONResponse(
            {"error": f"failed to parse docx: {e}"},
            status_code=400
        )

    text_parts = []
    for p in doc.paragraphs:
        if p.text.strip():
            text_parts.append(p.text)

    full_text = "\n".join(text_parts)

    return JSONResponse({
        "mode": "text",
        "text": full_text
    })


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)