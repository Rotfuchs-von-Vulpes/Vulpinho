import sys
import io
import cairosvg
from lxml import etree

def main():
    if len(sys.argv) < 2:
        sys.stderr.write("No args.\n")
        sys.exit(1)
        return
    
    try:
        svg_code = sys.argv[1]
        parser = etree.XMLParser(recover=False)
        etree.fromstring(svg_code.encode('utf-8'), parser=parser)
        
        png_data = cairosvg.svg2png(bytestring=svg_code.encode('utf-8'))
        sys.stdout.buffer.write(io.BytesIO(png_data).getvalue())
        sys.exit(0)
        return
    except etree.XMLSyntaxError as e:
        sys.stderr.write(f"Erro de Sintaxe XML (SVG Inválido): {e}")
        sys.exit(2)
        return
    except Exception as e:
        sys.stderr.write(f"Erro no CairoSVG: {e}")
        sys.exit(3)
        return

if __name__ == "__main__":
    main()