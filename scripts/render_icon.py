from pathlib import Path
from PyQt5.QtCore import QSize
from PyQt5.QtGui import QImage, QPainter
from PyQt5.QtSvg import QSvgRenderer

root = Path(__file__).resolve().parent.parent
renderer = QSvgRenderer(str(root / "app" / "icon.svg"))
if not renderer.isValid():
    raise SystemExit("invalid SVG")
image = QImage(QSize(256, 256), QImage.Format_ARGB32)
image.fill(0)
painter = QPainter(image)
renderer.render(painter)
painter.end()
if not image.save(str(root / "app" / "icon.png"), "PNG"):
    raise SystemExit("failed to save PNG")
