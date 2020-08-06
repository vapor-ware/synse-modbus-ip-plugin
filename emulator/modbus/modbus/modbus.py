#!/usr/bin/env python3
"""modbus main. A simple modbus server, here to test bulk reads."""


import time
import traceback

from pyModbusTCP.server import DataBank, ModbusServer


def main():
    """ Main method to run the pyModbusTCP server for the test modbus endpoint.
    """
    try:
        print('modbus main start')
        # Set holding register data to their address.
        # Set coils to address % 3 == 0.
        # This is a high number, but required for spliting a ModbusRead for coils.
        # TODO: ^^^Test the above. (with test-endpoints)
        for i in range(0x4000):
            print('Setting word {} to {}'.format(i, i))
            print('Setting bits {} to {}'.format(i, i % 3 == 0))
            DataBank.set_words(i, [i])
            DataBank.set_bits(i, [i % 3 == 0])

        server = ModbusServer(host='', port=1502, no_block=True)
        server.start()
        print('started modbus server')
        global STOP_SERVER
        while not STOP_SERVER:
            time.sleep(.01)
        server.stop()
        print('stopped modbus server')
    except Exception:
        traceback.print_exc()


if __name__ == '__main__':
    main()

