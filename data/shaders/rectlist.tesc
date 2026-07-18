#version 450
// 3 input patch vertices per rect; 4 output control points for the quad.
layout(vertices = 4) out;

layout(location = 0) in vec4 param_in_0[];
layout(location = 1) in vec4 param_in_1[];
layout(location = 2) in vec4 param_in_2[];
layout(location = 3) in vec4 param_in_3[];
layout(location = 4) in vec4 param_in_4[];
layout(location = 5) in vec4 param_in_5[];
layout(location = 6) in vec4 param_in_6[];
layout(location = 7) in vec4 param_in_7[];
layout(location = 8) in vec4 param_in_8[];
layout(location = 9) in vec4 param_in_9[];
layout(location = 10) in vec4 param_in_10[];
layout(location = 11) in vec4 param_in_11[];
layout(location = 12) in vec4 param_in_12[];
layout(location = 13) in vec4 param_in_13[];
layout(location = 14) in vec4 param_in_14[];
layout(location = 15) in vec4 param_in_15[];

layout(location = 0) out vec4 param_out_0[];
layout(location = 1) out vec4 param_out_1[];
layout(location = 2) out vec4 param_out_2[];
layout(location = 3) out vec4 param_out_3[];
layout(location = 4) out vec4 param_out_4[];
layout(location = 5) out vec4 param_out_5[];
layout(location = 6) out vec4 param_out_6[];
layout(location = 7) out vec4 param_out_7[];
layout(location = 8) out vec4 param_out_8[];
layout(location = 9) out vec4 param_out_9[];
layout(location = 10) out vec4 param_out_10[];
layout(location = 11) out vec4 param_out_11[];
layout(location = 12) out vec4 param_out_12[];
layout(location = 13) out vec4 param_out_13[];
layout(location = 14) out vec4 param_out_14[];
layout(location = 15) out vec4 param_out_15[];

vec4 interpolate(vec4 v0, vec4 v1, vec4 v2, float bary_coord[3]) {
    return v0 * bary_coord[0] + v1 * bary_coord[1] + v2 * bary_coord[2];
}

void emitParam(out vec4 param_out[4], vec4 param_in[3], int index, bool invocation_3, float bary_coord[3]) {
    vec4 param3 = interpolate(param_in[0], param_in[1], param_in[2], bary_coord);
    param_out[gl_InvocationID] = invocation_3 ? param3 : param_in[index];
}

void main() {
    gl_TessLevelOuter[0] = 1.0;
    gl_TessLevelOuter[1] = 1.0;
    gl_TessLevelOuter[2] = 1.0;
    gl_TessLevelOuter[3] = 1.0;
    gl_TessLevelInner[0] = 1.0;
    gl_TessLevelInner[1] = 1.0;

    vec4 pos[3];
    pos[0] = gl_in[0].gl_Position;
    pos[1] = gl_in[1].gl_Position;
    pos[2] = gl_in[2].gl_Position;

    bvec2 point_coord_equal[3];
    for (int i = 0; i < 3; i++) {
        point_coord_equal[i] = equal(pos[i].xy, pos[(i + 1) % 3].xy);
    }

    float bary_coord[3];
    bool is_edge_vertex[3];
    for (int i = 0; i < 3; i++) {
        bool xy_equal = point_coord_equal[i].x && point_coord_equal[(i + 2) % 3].y;
        bool yx_equal = point_coord_equal[i].y && point_coord_equal[(i + 2) % 3].x;
        is_edge_vertex[i] = xy_equal || yx_equal;
        bary_coord[i] = is_edge_vertex[i] ? -1.0 : 1.0;
    }

    int vertex_index = is_edge_vertex[2] ? 2 : 0;
    if (is_edge_vertex[1]) {
        vertex_index = 1;
    }

    int index = (vertex_index + gl_InvocationID) % 3;
    bool invocation_3 = gl_InvocationID == 3;

    vec4 pos3 = interpolate(pos[0], pos[1], pos[2], bary_coord);
    gl_out[gl_InvocationID].gl_Position = invocation_3 ? pos3 : gl_in[index].gl_Position;

    vec4 in0[3] = vec4[3](param_in_0[0], param_in_0[1], param_in_0[2]);
    vec4 in1[3] = vec4[3](param_in_1[0], param_in_1[1], param_in_1[2]);
    vec4 in2[3] = vec4[3](param_in_2[0], param_in_2[1], param_in_2[2]);
    vec4 in3[3] = vec4[3](param_in_3[0], param_in_3[1], param_in_3[2]);
    vec4 in4[3] = vec4[3](param_in_4[0], param_in_4[1], param_in_4[2]);
    vec4 in5[3] = vec4[3](param_in_5[0], param_in_5[1], param_in_5[2]);
    vec4 in6[3] = vec4[3](param_in_6[0], param_in_6[1], param_in_6[2]);
    vec4 in7[3] = vec4[3](param_in_7[0], param_in_7[1], param_in_7[2]);
    vec4 in8[3] = vec4[3](param_in_8[0], param_in_8[1], param_in_8[2]);
    vec4 in9[3] = vec4[3](param_in_9[0], param_in_9[1], param_in_9[2]);
    vec4 in10[3] = vec4[3](param_in_10[0], param_in_10[1], param_in_10[2]);
    vec4 in11[3] = vec4[3](param_in_11[0], param_in_11[1], param_in_11[2]);
    vec4 in12[3] = vec4[3](param_in_12[0], param_in_12[1], param_in_12[2]);
    vec4 in13[3] = vec4[3](param_in_13[0], param_in_13[1], param_in_13[2]);
    vec4 in14[3] = vec4[3](param_in_14[0], param_in_14[1], param_in_14[2]);
    vec4 in15[3] = vec4[3](param_in_15[0], param_in_15[1], param_in_15[2]);

    param_out_0[gl_InvocationID] = invocation_3 ? interpolate(in0[0], in0[1], in0[2], bary_coord) : in0[index];
    param_out_1[gl_InvocationID] = invocation_3 ? interpolate(in1[0], in1[1], in1[2], bary_coord) : in1[index];
    param_out_2[gl_InvocationID] = invocation_3 ? interpolate(in2[0], in2[1], in2[2], bary_coord) : in2[index];
    param_out_3[gl_InvocationID] = invocation_3 ? interpolate(in3[0], in3[1], in3[2], bary_coord) : in3[index];
    param_out_4[gl_InvocationID] = invocation_3 ? interpolate(in4[0], in4[1], in4[2], bary_coord) : in4[index];
    param_out_5[gl_InvocationID] = invocation_3 ? interpolate(in5[0], in5[1], in5[2], bary_coord) : in5[index];
    param_out_6[gl_InvocationID] = invocation_3 ? interpolate(in6[0], in6[1], in6[2], bary_coord) : in6[index];
    param_out_7[gl_InvocationID] = invocation_3 ? interpolate(in7[0], in7[1], in7[2], bary_coord) : in7[index];
    param_out_8[gl_InvocationID] = invocation_3 ? interpolate(in8[0], in8[1], in8[2], bary_coord) : in8[index];
    param_out_9[gl_InvocationID] = invocation_3 ? interpolate(in9[0], in9[1], in9[2], bary_coord) : in9[index];
    param_out_10[gl_InvocationID] = invocation_3 ? interpolate(in10[0], in10[1], in10[2], bary_coord) : in10[index];
    param_out_11[gl_InvocationID] = invocation_3 ? interpolate(in11[0], in11[1], in11[2], bary_coord) : in11[index];
    param_out_12[gl_InvocationID] = invocation_3 ? interpolate(in12[0], in12[1], in12[2], bary_coord) : in12[index];
    param_out_13[gl_InvocationID] = invocation_3 ? interpolate(in13[0], in13[1], in13[2], bary_coord) : in13[index];
    param_out_14[gl_InvocationID] = invocation_3 ? interpolate(in14[0], in14[1], in14[2], bary_coord) : in14[index];
    param_out_15[gl_InvocationID] = invocation_3 ? interpolate(in15[0], in15[1], in15[2], bary_coord) : in15[index];
}