import os
import sys
from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
from flask_migrate import Migrate
from sqlalchemy.exc import OperationalError
from src.infrastructure.db.models import db, Producto, Area, Receta, Ingrediente, MovimientoInventario, Venta, InventarioDiario, ModeloIPV
from src.presentation.controllers import producto_controller
from src.presentation.controllers import area_controller
from src.presentation.controllers import receta_controller
from src.presentation.controllers import venta_controller
from src.presentation.controllers import inventario_diario_controller
from src.presentation.controllers import historial_controller
from src.infrastructure.container import Container
import webbrowser
from threading import Timer

from datetime import datetime

def create_app() -> Flask:
    # Cargar variables de entorno
    dotenv_path = os.path.join(os.path.dirname(__file__), 'env', '.env')
    load_dotenv(dotenv_path=dotenv_path)
    
    print(f"App reloaded at: {datetime.now()}")

    container = Container()
    container.wire(modules=[
        producto_controller, 
        area_controller, 
        receta_controller, 
        venta_controller,
        inventario_diario_controller,
        historial_controller
    ])
    
    if getattr(sys, 'frozen', False):
        # Running as a PyInstaller bundle
        static_folder = os.path.join(sys._MEIPASS, 'frontend/dist')
    else:
        # Running as a normal script
        static_folder = '../frontend/dist'

    app = Flask(__name__, static_folder=static_folder, static_url_path='/')
    app.container = container
    app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DB_URI', 'sqlite:///inventario.db')
    app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
    
    # Configurar CORS
    CORS(app, resources={r"/api/*": {"origins": "*"}})

    app.register_blueprint(producto_controller.producto_bp)
    app.register_blueprint(area_controller.area_bp)
    app.register_blueprint(receta_controller.receta_bp)
    app.register_blueprint(venta_controller.venta_bp)
    app.register_blueprint(inventario_diario_controller.inventario_diario_bp)
    app.register_blueprint(historial_controller.historial_bp)

    db.init_app(app)
    migrate = Migrate(app, db)
    
    container.db_session.override(db.session)

    with app.app_context():
        try:
            db.create_all()
        except OperationalError as e:
            app.logger.warning(
                "Could not create database tables."
                " This can be ignored if the tables already exist."
                f" Error: {e}"
            )

    @app.route('/', defaults={'path': ''})
    @app.route('/<path:path>')
    def catch_all(path):
        return app.send_static_file("index.html")

    return app

app = create_app()

if __name__ == '__main__':
    import threading
    from waitress import serve
    from pystray import Icon as TrayIcon, Menu, MenuItem
    from PIL import Image
    import signal
    from src.icon import get_icon_image

    # --- Configuration ---
    PORT = int(os.environ.get("PORT", 5000))
    URL = f"http://127.0.0.1:{PORT}"

    # --- Server Control ---
    server_thread = None
    tray_icon = None

    def run_server():
        """Function to run the Waitress server."""
        print(f"Starting server on {URL}")
        serve(app, host="0.0.0.0", port=PORT)

    def open_browser():
        """Open the web browser to the application URL."""
        webbrowser.open(URL)

    def exit_action(icon, item):
        """Function to stop the server and exit the application."""
        print("Stopping server...")
        # This is a bit abrupt but necessary to stop waitress from a different thread
        os.kill(os.getpid(), signal.SIGTERM)
        icon.stop()

    # --- System Tray Icon Setup ---
    def setup_tray():
        """Sets up and runs the system tray icon."""
        global tray_icon
        
        image = get_icon_image()

        menu = Menu(
            MenuItem('Open', open_browser, default=True),
            MenuItem('Exit', exit_action)
        )
        tray_icon = TrayIcon("Inventario", image, "Inventario Restaurante", menu)
        
        # Open browser shortly after starting
        Timer(1, open_browser).start()
        
        tray_icon.run()

    # --- Main Execution ---
    # Start the server in a background thread
    server_thread = threading.Thread(target=run_server)
    server_thread.daemon = True
    server_thread.start()

    # Run the system tray icon in the main thread
    setup_tray()
