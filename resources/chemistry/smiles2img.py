import sys
import io
import re
import cairosvg
from io import BytesIO
from io import StringIO
from rdkit import Chem
from rdkit.Chem import Draw
from rdkit.Chem.Draw import rdMolDraw2D

def parseSmiles(smiles):
    sio = StringIO()
    save_stderr = sys.stderr
    sys.stderr = sio
    
    try:
        mol = Chem.MolFromSmiles(smiles)
        
        if mol is None:
            return None, sio.getvalue().strip()
        return mol, None
        
    finally:
        sys.stderr = save_stderr

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
    
    # Generate the molecule image
    drawer = rdMolDraw2D.MolDraw2DSVG(256, 256)
    opts = drawer.drawOptions()
    opts.clearBackground = False
    opts.scaleBondWidth = True
    opts.bondLineWidth = 5
    # opts.baseFontSize = 10
    drawer.DrawMolecule(mol)
    drawer.FinishDrawing()

    svg = drawer.GetDrawingText()
    svg = re.sub(r"#000000", "#FFFFFF", svg, flags=re.IGNORECASE)
    svg = re.sub(r"black", "white", svg, flags=re.IGNORECASE)
    png_data = cairosvg.svg2png(bytestring=svg.encode("utf-8"), scale=2)
    sys.stdout.buffer.write(io.BytesIO(png_data).getvalue())
    sys.exit(0)
    return

if __name__ == "__main__":
    main()