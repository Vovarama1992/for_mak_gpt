from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from docx import Document
from PIL import Image, ImageDraw, ImageFont
import io, base64, textwrap
import uvicorn

app = FastAPI()

@app.post("/convert")
async def convert(request: Request):
    raw = await request.body()

    # читаем DOCX
    doc = Document(io.BytesIO(raw))

    # собираем текст
    full_text = "\n".join(
        p.text for p in doc.paragraphs
        if p.text.strip()
    )

    # КОНСТАНТЫ — оптимальны, можешь не трогать
    PAGE_WIDTH = 1600
    PAGE_HEIGHT = 2200
    LEFT_PAD = 80
    TOP_PAD = 80
    LINE_SPACING = 60    # расстояние между строками
    FONT_SIZE = 36

    # грузим шрифт (встроенный PIL)
    font = ImageFont.truetype("DejaVuSans.ttf", FONT_SIZE)

    # вручную переносим строки по ширине страницы
    # чтобы текст не вылезал за край
    wrapped_lines = []
    for line in full_text.split("\n"):
        wrapped_lines.extend(
            textwrap.wrap(line, width=60) or [""]
        )

    # считаем сколько строк помещается на страницу
    LINES_PER_PAGE = (PAGE_HEIGHT - TOP_PAD * 2) // LINE_SPACING

    # режем на страницы
    pages_text = [
        wrapped_lines[i:i+LINES_PER_PAGE]
        for i in range(0, len(wrapped_lines), LINES_PER_PAGE)
    ]

    pages = []

    for page_index, lines in enumerate(pages_text, start=1):
        # создаём белый лист
        img = Image.new("RGB", (PAGE_WIDTH, PAGE_HEIGHT), "white")
        d = ImageDraw.Draw(img)

        y = TOP_PAD
        for line in lines:
            d.text((LEFT_PAD, y), line, font=font, fill="black")
            y += LINE_SPACING

        # кодируем
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=90)
        b64 = base64.b64encode(buf.getvalue()).decode()

        pages.append({
            "file_name": f"page-{page_index}.jpg",
            "mime": "image/jpeg",
            "base64": b64,
        })

    return JSONResponse({"pages": pages})


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)