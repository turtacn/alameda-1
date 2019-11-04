from setuptools import setup

# Ensure we're in the proper directory whether or not we're being used by pip.
#os.chdir(os.path.dirname(os.path.abspath(__file__)))

version='0.1'

with open('README.md', 'r') as f:
    readme = f.read()


with open('LICENSE', 'r') as f:
    license = f.read()

INSTALL_REQUIRES = (
    'protobuf>=3.9.0',
    'grpcio>=1.22.0',
    'grpcio-tools>=1.22.0',
    'googleapis-common-protos>=1.6.0',
)

setup(
    name='alameda-api',
    version=version,
    description='Alameda API interfaces',
    long_description=readme,
    long_description_content_type="text/markdown",
    author='ProphetStor Inc.',
    author_email='support@prophetstor.com',
    urls='https://github.com/containers-ai/api',
    license=license,
    packages=['alameda_api.v1alpha1.ai_service',
              'alameda_api.v1alpha1.operator',
              'alameda_api.v1alpha1.datahub',
              'alameda_api.v1alpha1.datahub.common',
              'alameda_api.v1alpha1.datahub.events',
              'alameda_api.v1alpha1.datahub.gpu',
              'alameda_api.v1alpha1.datahub.licenses',
              'alameda_api.v1alpha1.datahub.metrics',
              'alameda_api.v1alpha1.datahub.plannings',
              'alameda_api.v1alpha1.datahub.predictions',
              'alameda_api.v1alpha1.datahub.rawdata',
              'alameda_api.v1alpha1.datahub.recommendations',
              'alameda_api.v1alpha1.datahub.resources',
              'alameda_api.v1alpha1.datahub.scores',
              'alameda_api.v1alpha1.datahub.weavescope',
              'common'],
    package_dir={
        'alameda_api.v1alpha1.ai_service': 'alameda_api/v1alpha1/ai_service',
        'alameda_api.v1alpha1.operator': 'alameda_api/v1alpha1/operator',
        'alameda_api.v1alpha1.datahub': 'alameda_api/v1alpha1/datahub',
        'alameda_api.v1alpha1.datahub.common': 'alameda_api/v1alpha1/datahub/common',
        'alameda_api.v1alpha1.datahub.events': 'alameda_api/v1alpha1/datahub/events',
        'alameda_api.v1alpha1.datahub.gpu': 'alameda_api/v1alpha1/datahub/gpu',
        'alameda_api.v1alpha1.datahub.licenses': 'alameda_api/v1alpha1/datahub/licenses',
        'alameda_api.v1alpha1.datahub.metrics': 'alameda_api/v1alpha1/datahub/metrics',
        'alameda_api.v1alpha1.datahub.plannings': 'alameda_api/v1alpha1/datahub/plannings',
        'alameda_api.v1alpha1.datahub.predictions': 'alameda_api/v1alpha1/datahub/predictions',
        'alameda_api.v1alpha1.datahub.rawdata': 'alameda_api/v1alpha1/datahub/rawdata',
        'alameda_api.v1alpha1.datahub.recommendations': 'alameda_api/v1alpha1/datahub/recommendations',
        'alameda_api.v1alpha1.datahub.resources': 'alameda_api/v1alpha1/datahub/resources',
        'alameda_api.v1alpha1.datahub.scores': 'alameda_api/v1alpha1/datahub/scores',
        'alameda_api.v1alpha1.datahub.weavescope': 'alameda_api/v1alpha1/datahub/weavescope',
        'common': 'common',
    },
    install_requires=INSTALL_REQUIRES,
    zip_safe=False
)

