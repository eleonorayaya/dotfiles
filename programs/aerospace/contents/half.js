ObjC.import('AppKit');

function loadConfig() {
  const path = $.NSString.stringWithString('~/.config/aerospace/half.json').stringByExpandingTildeInPath;
  const text = $.NSString.stringWithContentsOfFileEncodingError(path, $.NSUTF8StringEncoding, null);
  return JSON.parse(text.js);
}

function screenFor(x, y) {
  const screens = $.NSScreen.screens;
  const primaryHeight = screens.objectAtIndex(0).frame.size.height;
  let match = null;
  for (let i = 0; i < screens.count; i++) {
    const screen = screens.objectAtIndex(i);
    const f = screen.frame;
    const top = primaryHeight - (f.origin.y + f.size.height);
    if (x >= f.origin.x && x < f.origin.x + f.size.width && y >= top && y < top + f.size.height) {
      match = screen;
    }
  }
  const screen = match || screens.objectAtIndex(0);
  const v = screen.visibleFrame;
  return {
    name: screen.localizedName.js,
    x: v.origin.x,
    y: primaryHeight - (v.origin.y + v.size.height),
    width: v.size.width,
    height: v.size.height,
  };
}

function run(argv) {
  const side = argv[0] === 'right' ? 'right' : 'left';
  const events = Application('System Events');
  const proc = events.processes.whose({ frontmost: true })[0];
  const win = proc.attributes.byName('AXFocusedWindow').value();
  const [wx, wy] = win.position();
  const [ww, wh] = win.size();

  const config = loadConfig();
  const screen = screenFor(wx + ww / 2, wy + wh / 2);
  const override = config.override;
  const g = override && new RegExp(override.pattern, 'i').test(screen.name) ? override.gaps : config.gaps;
  const width = Math.floor((screen.width - g.left - g.right - g.inner) / 2);
  const height = screen.height - g.top - g.bottom;
  const x = screen.x + g.left + (side === 'right' ? width + g.inner : 0);
  const y = screen.y + g.top;

  win.position = [x, y];
  win.size = [width, height];
  win.position = [x, y];
}
