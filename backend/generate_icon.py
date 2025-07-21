import base64

try:
    with open("frontend/public/icon.png", "rb") as image_file:
        encoded_string = base64.b64encode(image_file.read()).decode('utf-8')

    with open("backend/src/icon.py", "w") as f:
        f.write("# This file is auto-generated. Do not edit.\n")
        f.write("import base64\n")
        f.write("from PIL import Image\n")
        f.write("import io\n\n")
        f.write(f"ICON_B64 = '{encoded_string}'\n\n")
        f.write("def get_icon_image():\n")
        f.write("    image_data = base64.b64decode(ICON_B64)\n")
        f.write("    image = Image.open(io.BytesIO(image_data))\n")
        f.write("    return image\n")

    print("icon.py generated successfully at backend/src/icon.py")

except FileNotFoundError:
    print("Error: frontend/public/icon.png not found. Please ensure the icon file exists.")
except Exception as e:
    print(f"An error occurred: {e}")
