import serial
import random
import time
import sys, traceback
from google.cloud import firestore

leds = 512*8

db = firestore.Client(project="mofilabs-next-demo-02")

def randomColors():
    colorStr = []

    for i in range(leds):
        r = random.randint(0,50)
        g = random.randint(0,50)
        b = random.randint(0,50)
        colorStr.append(r)
        colorStr.append(g)
        colorStr.append(b)

    return bytearray(colorStr)



def clear():
    colorStr = []
    for i in range(leds):
        r = 0
        g = 0
        b = 0
        colorStr.append(r)
        colorStr.append(g)
        colorStr.append(b)
    return bytearray(colorStr)

def get_byte_data():
    # Reference the specific document
    doc_ref = db.collection('led').document('data')

    # Fetch the document
    doc = doc_ref.get()

    # Check if the document exists
    if doc.exists:
        # Fetch the byte data
        byte_data = doc.get('data')
        if byte_data:
            return byte_data
        else:
            print("Field 'data' does not exist or is None.")
            return None
    else:
        print("Document does not exist.")
        return None

def main():
    try:
        ser = serial.Serial('/dev/ttyACM0', 9600)
        doc_ref = db.collection("led").document("data")
        while True:
            color = get_byte_data()
            ser.write(color)
            time.sleep(1)

    except KeyboardInterrupt:
        print("Shutdown requested...exiting")
        color = clear()
        ser.write(color)
        time.sleep(1)
        ser.close()
        time.sleep(1)
    except Exception:
        traceback.print_exc(file=sys.stdout)
    sys.exit(0)



if __name__ == '__main__':
    main()