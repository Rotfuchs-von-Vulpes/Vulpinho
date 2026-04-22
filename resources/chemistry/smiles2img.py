import sys
import io
import re
import cairosvg
from PIL import Image
from io import BytesIO
from io import StringIO
from rdkit import Chem
from rdkit.Chem import Draw
from rdkit.Chem.Draw import rdMolDraw2D

def main():
    if len(sys.argv) < 2:
        sys.stderr.write("No args.\n")
        sys.exit(1)
        return
    
    smiles = sys.argv[1]
    mol = Chem.MolFromSmiles(smiles)
    if mol is None:
        sys.exit(2)
        return
    
    drawer = rdMolDraw2D.MolDraw2DSVG(256, 256)
    opts = drawer.drawOptions()
    opts.clearBackground = False
    opts.scaleBondWidth = True
    opts.bondLineWidth = 5
    drawer.DrawMolecule(mol)
    drawer.FinishDrawing()

    svg = drawer.GetDrawingText()
    svg = re.sub(r"#000000", "#FFFFFF", svg, flags=re.IGNORECASE)
    svg = re.sub(r"black", "white", svg, flags=re.IGNORECASE)
    png_data = cairosvg.svg2png(bytestring=svg.encode("utf-8"), scale=2)
    img = Image.open(io.BytesIO(png_data))
    bbox = img.getbbox()
    if bbox:
        margin = 50
        width, height = img.size
        left = max(0, bbox[0] - margin)
        upper = max(0, bbox[1] - margin)
        right = min(width, bbox[2] + margin)
        lower = min(height, bbox[3] + margin)
        cropped_image = img.crop((left, upper, right, lower))

        buf = io.BytesIO()
        cropped_image.save(buf, format='PNG')
        sys.stdout.buffer.write(buf.getvalue())
        sys.exit(0)
    else:
        sys.stdout.buffer.write(io.BytesIO(png_data).getvalue())
        sys.exit(0)
    return

if __name__ == "__main__":
    main()