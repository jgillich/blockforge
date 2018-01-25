#ifdef cl_amd_media_ops
#pragma OPENCL EXTENSION cl_amd_media_ops : enable
#else
/* taken from https://www.khronos.org/registry/OpenCL/extensions/amd/cl_amd_media_ops.txt
 * Build-in Function
 *     uintn  amd_bitalign (uintn src0, uintn src1, uintn src2)
 *   Description
 *     dst.s0 =  (uint) (((((long)src0.s0) << 32) | (long)src1.s0) >> (src2.s0 & 31))
 *     similar operation applied to other components of the vectors.
 *
 * The implemented function is modified because the last is in our case always a scalar.
 * We can ignore the bitwise AND operation.
 */
inline uint2 amd_bitalign( const uint2 src0, const uint2 src1, const uint src2)
{
	uint2 result;
	result.s0 =  (uint) (((((long)src0.s0) << 32) | (long)src1.s0) >> (src2));
	result.s1 =  (uint) (((((long)src0.s1) << 32) | (long)src1.s1) >> (src2));
	return result;
}
#endif

#ifdef cl_amd_media_ops2
#pragma OPENCL EXTENSION cl_amd_media_ops2 : enable
#else
/* taken from: https://www.khronos.org/registry/OpenCL/extensions/amd/cl_amd_media_ops2.txt
 *     Built-in Function:
 *     uintn amd_bfe (uintn src0, uintn src1, uintn src2)
 *   Description
 *     NOTE: operator >> below represent logical right shift
 *     offset = src1.s0 & 31;
 *     width = src2.s0 & 31;
 *     if width = 0
 *         dst.s0 = 0;
 *     else if (offset + width) < 32
 *         dst.s0 = (src0.s0 << (32 - offset - width)) >> (32 - width);
 *     else
 *         dst.s0 = src0.s0 >> offset;
 *     similar operation applied to other components of the vectors
 */
inline int amd_bfe(const uint src0, const uint offset, const uint width)
{
	/* casts are removed because we can implement everything as uint
	 * int offset = src1;
	 * int width = src2;
	 * remove check for edge case, this function is always called with
	 * `width==8`
	 * @code
	 *   if ( width == 0 )
	 *      return 0;
	 * @endcode
	 */
	if ( (offset + width) < 32u )
		return (src0 << (32u - offset - width)) >> (32u - width);

	return src0 >> offset;
}
#endif
