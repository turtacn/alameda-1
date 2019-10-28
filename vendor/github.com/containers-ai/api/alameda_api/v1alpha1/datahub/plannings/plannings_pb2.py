# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: alameda_api/v1alpha1/datahub/plannings/plannings.proto

import sys
_b=sys.version_info[0]<3 and (lambda x:x) or (lambda x:x.encode('latin1'))
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from alameda_api.v1alpha1.datahub.common import metrics_pb2 as alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2
from alameda_api.v1alpha1.datahub.plannings import types_pb2 as alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2
from alameda_api.v1alpha1.datahub.resources import metadata_pb2 as alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_metadata__pb2
from alameda_api.v1alpha1.datahub.resources import policies_pb2 as alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_policies__pb2
from alameda_api.v1alpha1.datahub.resources import types_pb2 as alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_types__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='alameda_api/v1alpha1/datahub/plannings/plannings.proto',
  package='containersai.alameda.v1alpha1.datahub.plannings',
  syntax='proto3',
  serialized_options=_b('ZCgithub.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings'),
  serialized_pb=_b('\n6alameda_api/v1alpha1/datahub/plannings/plannings.proto\x12/containersai.alameda.v1alpha1.datahub.plannings\x1a\x31\x61lameda_api/v1alpha1/datahub/common/metrics.proto\x1a\x32\x61lameda_api/v1alpha1/datahub/plannings/types.proto\x1a\x35\x61lameda_api/v1alpha1/datahub/resources/metadata.proto\x1a\x35\x61lameda_api/v1alpha1/datahub/resources/policies.proto\x1a\x32\x61lameda_api/v1alpha1/datahub/resources/types.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\x81\x03\n\x11\x43ontainerPlanning\x12\x0c\n\x04name\x18\x01 \x01(\t\x12Q\n\x0flimit_plannings\x18\x02 \x03(\x0b\x32\x38.containersai.alameda.v1alpha1.datahub.common.MetricData\x12S\n\x11request_plannings\x18\x03 \x03(\x0b\x32\x38.containersai.alameda.v1alpha1.datahub.common.MetricData\x12Y\n\x17initial_limit_plannings\x18\x04 \x03(\x0b\x32\x38.containersai.alameda.v1alpha1.datahub.common.MetricData\x12[\n\x19initial_request_plannings\x18\x05 \x03(\x0b\x32\x38.containersai.alameda.v1alpha1.datahub.common.MetricData\"\xf6\x04\n\x0bPodPlanning\x12T\n\rplanning_type\x18\x01 \x01(\x0e\x32=.containersai.alameda.v1alpha1.datahub.plannings.PlanningType\x12X\n\x0fnamespaced_name\x18\x02 \x01(\x0b\x32?.containersai.alameda.v1alpha1.datahub.resources.NamespacedName\x12\x1a\n\x12\x61pply_planning_now\x18\x03 \x01(\x08\x12[\n\x11\x61ssign_pod_policy\x18\x04 \x01(\x0b\x32@.containersai.alameda.v1alpha1.datahub.resources.AssignPodPolicy\x12_\n\x13\x63ontainer_plannings\x18\x05 \x03(\x0b\x32\x42.containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning\x12.\n\nstart_time\x18\x06 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12,\n\x08\x65nd_time\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12V\n\x0etop_controller\x18\x08 \x01(\x0b\x32>.containersai.alameda.v1alpha1.datahub.resources.TopController\x12\x13\n\x0bplanning_id\x18\t \x01(\t\x12\x12\n\ntotal_cost\x18\n \x01(\x01\"\x9d\x03\n\x12\x43ontrollerPlanning\x12T\n\rplanning_type\x18\x01 \x01(\x0e\x32=.containersai.alameda.v1alpha1.datahub.plannings.PlanningType\x12\x62\n\x11\x63tl_planning_type\x18\x02 \x01(\x0e\x32G.containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanningType\x12\x62\n\x11\x63tl_planning_spec\x18\x03 \x01(\x0b\x32G.containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanningSpec\x12i\n\x15\x63tl_planning_spec_k8s\x18\x04 \x01(\x0b\x32J.containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanningSpecK8sBEZCgithub.com/containers-ai/api/alameda_api/v1alpha1/datahub/planningsb\x06proto3')
  ,
  dependencies=[alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2.DESCRIPTOR,alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2.DESCRIPTOR,alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_metadata__pb2.DESCRIPTOR,alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_policies__pb2.DESCRIPTOR,alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_types__pb2.DESCRIPTOR,google_dot_protobuf_dot_timestamp__pb2.DESCRIPTOR,])




_CONTAINERPLANNING = _descriptor.Descriptor(
  name='ContainerPlanning',
  full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='limit_plannings', full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning.limit_plannings', index=1,
      number=2, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='request_plannings', full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning.request_plannings', index=2,
      number=3, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='initial_limit_plannings', full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning.initial_limit_plannings', index=3,
      number=4, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='initial_request_plannings', full_name='containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning.initial_request_plannings', index=4,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=406,
  serialized_end=791,
)


_PODPLANNING = _descriptor.Descriptor(
  name='PodPlanning',
  full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='planning_type', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.planning_type', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='namespaced_name', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.namespaced_name', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='apply_planning_now', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.apply_planning_now', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='assign_pod_policy', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.assign_pod_policy', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='container_plannings', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.container_plannings', index=4,
      number=5, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='start_time', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.start_time', index=5,
      number=6, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='end_time', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.end_time', index=6,
      number=7, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='top_controller', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.top_controller', index=7,
      number=8, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='planning_id', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.planning_id', index=8,
      number=9, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=_b("").decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='total_cost', full_name='containersai.alameda.v1alpha1.datahub.plannings.PodPlanning.total_cost', index=9,
      number=10, type=1, cpp_type=5, label=1,
      has_default_value=False, default_value=float(0),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=794,
  serialized_end=1424,
)


_CONTROLLERPLANNING = _descriptor.Descriptor(
  name='ControllerPlanning',
  full_name='containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  fields=[
    _descriptor.FieldDescriptor(
      name='planning_type', full_name='containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning.planning_type', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='ctl_planning_type', full_name='containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning.ctl_planning_type', index=1,
      number=2, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='ctl_planning_spec', full_name='containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning.ctl_planning_spec', index=2,
      number=3, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
    _descriptor.FieldDescriptor(
      name='ctl_planning_spec_k8s', full_name='containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning.ctl_planning_spec_k8s', index=3,
      number=4, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1427,
  serialized_end=1840,
)

_CONTAINERPLANNING.fields_by_name['limit_plannings'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2._METRICDATA
_CONTAINERPLANNING.fields_by_name['request_plannings'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2._METRICDATA
_CONTAINERPLANNING.fields_by_name['initial_limit_plannings'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2._METRICDATA
_CONTAINERPLANNING.fields_by_name['initial_request_plannings'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_common_dot_metrics__pb2._METRICDATA
_PODPLANNING.fields_by_name['planning_type'].enum_type = alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2._PLANNINGTYPE
_PODPLANNING.fields_by_name['namespaced_name'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_metadata__pb2._NAMESPACEDNAME
_PODPLANNING.fields_by_name['assign_pod_policy'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_policies__pb2._ASSIGNPODPOLICY
_PODPLANNING.fields_by_name['container_plannings'].message_type = _CONTAINERPLANNING
_PODPLANNING.fields_by_name['start_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_PODPLANNING.fields_by_name['end_time'].message_type = google_dot_protobuf_dot_timestamp__pb2._TIMESTAMP
_PODPLANNING.fields_by_name['top_controller'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_resources_dot_types__pb2._TOPCONTROLLER
_CONTROLLERPLANNING.fields_by_name['planning_type'].enum_type = alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2._PLANNINGTYPE
_CONTROLLERPLANNING.fields_by_name['ctl_planning_type'].enum_type = alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2._CONTROLLERPLANNINGTYPE
_CONTROLLERPLANNING.fields_by_name['ctl_planning_spec'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2._CONTROLLERPLANNINGSPEC
_CONTROLLERPLANNING.fields_by_name['ctl_planning_spec_k8s'].message_type = alameda__api_dot_v1alpha1_dot_datahub_dot_plannings_dot_types__pb2._CONTROLLERPLANNINGSPECK8S
DESCRIPTOR.message_types_by_name['ContainerPlanning'] = _CONTAINERPLANNING
DESCRIPTOR.message_types_by_name['PodPlanning'] = _PODPLANNING
DESCRIPTOR.message_types_by_name['ControllerPlanning'] = _CONTROLLERPLANNING
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ContainerPlanning = _reflection.GeneratedProtocolMessageType('ContainerPlanning', (_message.Message,), {
  'DESCRIPTOR' : _CONTAINERPLANNING,
  '__module__' : 'alameda_api.v1alpha1.datahub.plannings.plannings_pb2'
  # @@protoc_insertion_point(class_scope:containersai.alameda.v1alpha1.datahub.plannings.ContainerPlanning)
  })
_sym_db.RegisterMessage(ContainerPlanning)

PodPlanning = _reflection.GeneratedProtocolMessageType('PodPlanning', (_message.Message,), {
  'DESCRIPTOR' : _PODPLANNING,
  '__module__' : 'alameda_api.v1alpha1.datahub.plannings.plannings_pb2'
  # @@protoc_insertion_point(class_scope:containersai.alameda.v1alpha1.datahub.plannings.PodPlanning)
  })
_sym_db.RegisterMessage(PodPlanning)

ControllerPlanning = _reflection.GeneratedProtocolMessageType('ControllerPlanning', (_message.Message,), {
  'DESCRIPTOR' : _CONTROLLERPLANNING,
  '__module__' : 'alameda_api.v1alpha1.datahub.plannings.plannings_pb2'
  # @@protoc_insertion_point(class_scope:containersai.alameda.v1alpha1.datahub.plannings.ControllerPlanning)
  })
_sym_db.RegisterMessage(ControllerPlanning)


DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)
