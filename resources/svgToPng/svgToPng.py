import sys
import io
import cairosvg

def main():
    svg_code = sys.stdin.read()

    if len(svg_code) == 0:
        sys.stderr.write("No data.\n")
        sys.exit(1)
        return
    
    try:
        png_data = cairosvg.svg2png(bytestring=svg_code)
        sys.stdout.buffer.write(io.BytesIO(png_data).getvalue())
        sys.exit(0)
        return
    except Exception as e:
        sys.stderr.write(f"Erro no CairoSVG: {e}")
        sys.exit(3)
        return

if __name__ == "__main__":
    main()