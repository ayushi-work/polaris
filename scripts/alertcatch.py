from http.server import HTTPServer, BaseHTTPRequestHandler
import json
n=[0]
class H(BaseHTTPRequestHandler):
    def do_POST(self):
        d=self.rfile.read(int(self.headers.get('Content-Length',0))).decode()
        n[0]+=1
        print('\n'+'='*20+f' ALERT {n[0]} '+'='*20)
        try: print(json.dumps(json.loads(d),indent=2))
        except: print(d)
        self.send_response(200); self.end_headers(); self.wfile.write(b'{"ok":true}')
    def log_message(self,*a): pass
print('listening on :8899',flush=True)
HTTPServer(('127.0.0.1',8899),H).serve_forever()
