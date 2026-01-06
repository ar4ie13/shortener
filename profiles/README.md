File: shortener
Type: inuse_space
Time: 2026-01-06 14:37:34 MSK
Showing nodes accounting for -2.92MB, 0.79% of 367.98MB total
Dropped 7 nodes (cum <= 1.84MB)
flat  flat%   sum%        cum   cum%
-8.24MB  2.24%  2.24%    -8.24MB  2.24%  github.com/ar4ie13/shortener/internal/repository/memory.(*MemStorage).SaveBatch
3.50MB  0.95%  1.29%     3.50MB  0.95%  encoding/json.(*decodeState).literalStore
2.50MB  0.68%  0.61%     2.50MB  0.68%  github.com/ar4ie13/shortener/internal/service.generateShortURL
-1.13MB  0.31%  0.91%    -1.13MB  0.31%  bytes.growSlice
1.02MB  0.28%  0.64%     1.02MB  0.28%  encoding/json.(*Decoder).refill
-1MB  0.27%  0.91%    -1.50MB  0.41%  runtime.allocm
0.88MB  0.24%  0.67%     1.38MB  0.38%  compress/flate.NewWriter (inline)
0.55MB  0.15%  0.52%    -0.42MB  0.11%  github.com/ar4ie13/shortener/internal/handlers.Handler.postURLJSONBatch
0.51MB  0.14%  0.38%     0.51MB  0.14%  encoding/xml.map.init.0
-0.50MB  0.14%  0.52%    -0.50MB  0.14%  reflect.growslice
-0.50MB  0.14%  0.66%    -0.50MB  0.14%  runtime.makeProfStackFP (inline)
-0.50MB  0.14%  0.79%    -0.50MB  0.14%  regexp/syntax.(*compiler).inst (inline)
-0.50MB  0.14%  0.93%    -0.50MB  0.14%  runtime.malg
0.50MB  0.14%  0.79%     0.50MB  0.14%  net/url.parse
-0.50MB  0.14%  0.93%    -0.50MB  0.14%  time.NewTicker
0.50MB  0.14%  0.79%     0.50MB  0.14%  compress/flate.(*compressor).initDeflate (inline)
0.50MB  0.14%  0.66%     0.50MB  0.14%  runtime.(*scavengerState).init
-0.50MB  0.14%  0.79%    -0.50MB  0.14%  runtime.acquireSudog
0     0%  0.79%    -1.13MB  0.31%  bytes.(*Buffer).Write
0     0%  0.79%    -1.13MB  0.31%  bytes.(*Buffer).grow
0     0%  0.79%     0.50MB  0.14%  compress/flate.(*compressor).init
0     0%  0.79%     1.38MB  0.38%  compress/gzip.(*Writer).Write
0     0%  0.79%     4.01MB  1.09%  encoding/json.(*Decoder).Decode
0     0%  0.79%     1.02MB  0.28%  encoding/json.(*Decoder).readValue
0     0%  0.79%    -0.25MB 0.068%  encoding/json.(*Encoder).Encode
0     0%  0.79%     3.50MB  0.95%  encoding/json.(*decodeState).array
0     0%  0.79%     3.50MB  0.95%  encoding/json.(*decodeState).object
0     0%  0.79%        3MB  0.81%  encoding/json.(*decodeState).unmarshal
0     0%  0.79%        3MB  0.81%  encoding/json.(*decodeState).value
0     0%  0.79%    -1.13MB  0.31%  encoding/json.(*encodeState).marshal
0     0%  0.79%    -1.13MB  0.31%  encoding/json.(*encodeState).reflectValue
0     0%  0.79%    -1.13MB  0.31%  encoding/json.addrTextMarshalerEncoder
0     0%  0.79%    -1.13MB  0.31%  encoding/json.arrayEncoder.encode
0     0%  0.79%    -1.13MB  0.31%  encoding/json.condAddrEncoder.encode
0     0%  0.79%    -1.13MB  0.31%  encoding/json.sliceEncoder.encode
0     0%  0.79%    -1.13MB  0.31%  encoding/json.structEncoder.encode
0     0%  0.79%     0.51MB  0.14%  encoding/xml.init
0     0%  0.79%     1.38MB  0.38%  github.com/ar4ie13/shortener/internal/handlers.(*compressWriter).Write
0     0%  0.79%    -0.42MB  0.11%  github.com/ar4ie13/shortener/internal/handlers.Handler.authMiddleware-fm.Handler.authMiddleware.func1
0     0%  0.79%    -0.42MB  0.11%  github.com/ar4ie13/shortener/internal/handlers.Handler.gzipMiddleware-fm.Handler.gzipMiddleware.func1
0     0%  0.79%    -0.42MB  0.11%  github.com/ar4ie13/shortener/internal/handlers.Handler.requestLogger-fm.Handler.requestLogger.func1
0     0%  0.79%    -5.24MB  1.42%  github.com/ar4ie13/shortener/internal/service.(*Service).SaveBatch
0     0%  0.79%    -0.50MB  0.14%  github.com/ar4ie13/shortener/internal/service.(*Service).deleteShortURLs
0     0%  0.79%    -0.42MB  0.11%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
0     0%  0.79%    -0.42MB  0.11%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
0     0%  0.79%    -0.42MB  0.11%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0%  0.79%    -0.42MB  0.11%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0%  0.79%    -0.50MB  0.14%  github.com/jackc/pgpassfile.init
0     0%  0.79%    -0.42MB  0.11%  net/http.(*conn).serve
0     0%  0.79%    -0.42MB  0.11%  net/http.HandlerFunc.ServeHTTP
0     0%  0.79%    -0.42MB  0.11%  net/http.serverHandler.ServeHTTP
0     0%  0.79%     0.50MB  0.14%  net/url.Parse
0     0%  0.79%    -0.50MB  0.14%  reflect.Value.Grow
0     0%  0.79%    -0.50MB  0.14%  reflect.Value.grow
0     0%  0.79%    -0.50MB  0.14%  regexp.Compile (inline)
0     0%  0.79%    -0.50MB  0.14%  regexp.MustCompile
0     0%  0.79%    -0.50MB  0.14%  regexp.compile
0     0%  0.79%    -0.50MB  0.14%  regexp/syntax.(*compiler).cap (inline)
0     0%  0.79%    -0.50MB  0.14%  regexp/syntax.(*compiler).compile
0     0%  0.79%    -0.50MB  0.14%  regexp/syntax.Compile
0     0%  0.79%     0.50MB  0.14%  runtime.bgscavenge
0     0%  0.79%    -0.50MB  0.14%  runtime.gcBgMarkWorker
0     0%  0.79%    -0.50MB  0.14%  runtime.gcMarkDone
0     0%  0.79%    -0.50MB  0.14%  runtime.mProfStackInit (inline)
0     0%  0.79%       -1MB  0.27%  runtime.mcall
0     0%  0.79%    -0.50MB  0.14%  runtime.mcommoninit
0     0%  0.79%    -0.50MB  0.14%  runtime.mstart
0     0%  0.79%    -0.50MB  0.14%  runtime.mstart0
0     0%  0.79%    -0.50MB  0.14%  runtime.mstart1
0     0%  0.79%    -1.50MB  0.41%  runtime.newm
0     0%  0.79%    -0.50MB  0.14%  runtime.newproc.func1
0     0%  0.79%    -0.50MB  0.14%  runtime.newproc1
0     0%  0.79%       -1MB  0.27%  runtime.park_m
0     0%  0.79%    -1.50MB  0.41%  runtime.resetspinning
0     0%  0.79%    -1.50MB  0.41%  runtime.schedule
0     0%  0.79%    -0.50MB  0.14%  runtime.semacquire (inline)
0     0%  0.79%    -0.50MB  0.14%  runtime.semacquire1
0     0%  0.79%    -1.50MB  0.41%  runtime.startm
0     0%  0.79%    -0.50MB  0.14%  runtime.systemstack
0     0%  0.79%    -1.50MB  0.41%  runtime.wakep