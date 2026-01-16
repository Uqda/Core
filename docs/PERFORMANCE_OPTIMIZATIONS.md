# Performance Optimizations for Latency Reduction

This document outlines code-level optimizations to reduce network latency in Uqda Network.

## Current Latency Issues

- **Observed latency**: 53-98ms
- **Target latency**: 20-50ms (with optimizations)

## Identified Optimization Points

### 1. Connection Timeouts

**Current Settings:**
- Handshake timeout: 6 seconds (`link.go:612`)
- TCP dial timeout: 5 seconds (`link_tcp.go:64`)
- WebSocket read/write timeout: 10 seconds (`link_ws.go:136-137`)
- UNIX socket timeout: 5 seconds (`link_unix.go:23`)

**Problem:**
- High timeouts cause unnecessary delays when connections fail
- Slow failure detection delays retry attempts
- But too short timeout breaks compatibility with older versions

**Optimization:**
- Reduce handshake timeout from 6s to 5s (balance between performance and compatibility)
- Reduce TCP dial timeout from 5s to 3s
- Reduce WebSocket timeouts from 10s to 5s
- Reduce UNIX socket timeout from 5s to 2s

**Expected Improvement:** 1-2ms reduction in connection establishment time (while maintaining backward compatibility)

---

### 2. Connection Backoff Algorithm

**Current Implementation:**
- Starts at 1 second (`time.Second << backoff`)
- Doubles each time (exponential backoff)
- Maximum backoff: `defaultBackoffLimit` (1h8m16s)

**Problem:**
- Initial backoff of 1 second is too long for fast retries
- Exponential growth is too aggressive for temporary failures

**Optimization:**
- Use adaptive backoff: start with 100ms, then 250ms, 500ms, 1s, 2s...
- Cap initial retries at 500ms for first 3 attempts
- Use jitter to prevent thundering herd

**Expected Improvement:** 500-900ms reduction in reconnection time after temporary failures

---

### 3. DNS Lookup Caching

**Current Implementation:**
- DNS lookup happens on every connection attempt (`link.go:702`)
- No caching of DNS results

**Problem:**
- DNS lookups add 10-50ms per connection attempt
- Repeated lookups for same hostname waste time

**Optimization:**
- Implement DNS result caching with TTL
- Cache successful lookups for 5 minutes
- Cache failed lookups for 30 seconds
- Use `net.LookupIP` with caching wrapper

**Expected Improvement:** 10-30ms reduction per connection attempt

---

### 4. Connection Pooling and Reuse

**Current Implementation:**
- Each connection attempt creates new dialer
- No connection reuse for same peer

**Problem:**
- TCP handshake overhead on every connection
- No benefit from TCP keepalive optimization

**Optimization:**
- Reuse dialers for same destination
- Implement connection warm-up for persistent peers
- Use TCP keepalive more aggressively

**Expected Improvement:** 5-15ms reduction in connection establishment

---

### 5. Buffer Management

**Current Implementation:**
- Uses sync.Pool for buffer allocation (`pool.go`)
- Allocates buffers based on MTU

**Problem:**
- Buffer allocation overhead for small packets
- No pre-allocation for common sizes

**Optimization:**
- Pre-allocate common buffer sizes (128, 512, 1500 bytes)
- Use buffer pools with size buckets
- Reduce allocation overhead

**Expected Improvement:** 1-3ms reduction per packet

---

### 6. Parallel Connection Attempts

**Current Implementation:**
- Sequential IP address attempts (`link.go:726`)
- Tries IPs one by one until success

**Problem:**
- Slow when first IP fails
- No parallel attempts for multiple IPs

**Optimization:**
- Attempt connections to multiple IPs in parallel
- Use first successful connection
- Cancel other attempts when one succeeds

**Expected Improvement:** 10-20ms reduction when multiple IPs available

---

### 7. QUIC Configuration

**Current Settings:**
- MaxIdleTimeout: 1 minute (`link_quic.go:45`)
- KeepAlivePeriod: 20 seconds (`link_quic.go:46`)

**Problem:**
- Long idle timeout may delay connection recovery
- KeepAlive period may be too frequent

**Optimization:**
- Reduce MaxIdleTimeout to 30 seconds
- Adjust KeepAlivePeriod to 10 seconds
- Optimize QUIC connection parameters

**Expected Improvement:** 5-10ms improvement for QUIC connections

---

### 8. Handshake Optimization

**Current Implementation:**
- Handshake includes metadata encoding/decoding
- Sequential write-then-read pattern

**Problem:**
- Handshake adds 5-15ms overhead
- Sequential operations add latency

**Optimization:**
- Optimize metadata encoding (reduce allocations)
- Use buffered I/O for handshake
- Parallel handshake operations where possible

**Expected Improvement:** 2-5ms reduction in handshake time

---

## Implementation Priority

### High Priority (Immediate Impact)
1. ✅ Reduce connection timeouts (handshake, dial)
2. ✅ Optimize connection backoff algorithm
3. ✅ Add DNS lookup caching

### Medium Priority (Moderate Impact)
4. ⏳ Implement parallel connection attempts
5. ⏳ Optimize buffer management
6. ⏳ Connection pooling and reuse

### Low Priority (Small Impact)
7. ⏳ QUIC configuration tuning
8. ⏳ Handshake optimization

---

## Expected Total Improvement

**Before Optimizations:**
- Connection establishment: 50-100ms
- Total latency: 53-98ms

**After Optimizations:**
- Connection establishment: 20-40ms
- Total latency: **20-50ms** (realistic target)

**Total Reduction:** 20-50ms improvement possible

---

## Testing Recommendations

1. **Benchmark current performance:**
   - Measure connection establishment time
   - Measure end-to-end latency
   - Track DNS lookup times

2. **Test each optimization:**
   - Apply one optimization at a time
   - Measure impact of each change
   - Verify no regressions

3. **Integration testing:**
   - Test with various network conditions
   - Test with different peer types (TCP, TLS, QUIC)
   - Test with multiple peers

4. **Production validation:**
   - Deploy to test environment
   - Monitor real-world latency
   - Collect metrics and feedback

---

## Notes

- Some latency is unavoidable (speed of light, encryption overhead)
- Focus on reducing unnecessary delays
- Balance between performance and reliability
- Monitor for any negative impacts on stability

